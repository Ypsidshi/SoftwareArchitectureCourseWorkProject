import http from 'k6/http';
import { check, sleep } from 'k6';

const AUTH_BASE = __ENV.AUTH_BASE || 'http://host.docker.internal:8081';
const DEAL_BASE = __ENV.DEAL_BASE || 'http://host.docker.internal:8082';

export const options = {
  stages: [
    { duration: '30s', target: 10 },
    { duration: '1m', target: 50 },
    { duration: '1m', target: 50 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_failed: ['rate<=0.05'],
    'http_req_duration{name:list_sanatoriums}': ['p(99)<1000'],
    'http_req_duration{name:create_booking}': ['p(99)<2000'],
  },
};

function isoDateUTC(dayOffset) {
  const d = new Date(Date.now() + dayOffset * 24 * 60 * 60 * 1000);
  const y = d.getUTCFullYear();
  const m = String(d.getUTCMonth() + 1).padStart(2, '0');
  const day = String(d.getUTCDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

export default function () {
  const unique = `${__VU}_${__ITER}_${Date.now()}`;
  const email = `k6_user_${unique}@example.com`;
  const password = 'Pass1234';

  const registerBody = JSON.stringify({
    email,
    password,
    full_name: `k6 user ${unique}`,
    role: 'client',
  });

  const registerRes = http.post(`${AUTH_BASE}/api/v1/users/register`, registerBody, {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: 'register_user' },
  });

  check(registerRes, {
    'register status is 200/201': (r) => r.status === 200 || r.status === 201,
  });

  const loginBody = JSON.stringify({ email, password });
  const loginRes = http.post(`${DEAL_BASE}/api/auth/login`, loginBody, {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: 'login_user' },
  });

  const loginOk = check(loginRes, {
    'login status is 200': (r) => r.status === 200,
    'login has access_token': (r) => {
      try {
        return !!r.json('access_token');
      } catch (_) {
        return false;
      }
    },
  });
  if (!loginOk) {
    sleep(0.2);
    return;
  }

  const token = loginRes.json('access_token');

  const listRes = http.get(`${DEAL_BASE}/api/sanatoriums?page=1&page_size=1`, {
    tags: { name: 'list_sanatoriums' },
  });

  const listOk = check(listRes, {
    'list status is 200': (r) => r.status === 200,
    'list has items': (r) => {
      try {
        const items = r.json('items');
        return Array.isArray(items) && items.length > 0;
      } catch (_) {
        return false;
      }
    },
  });
  if (!listOk) {
    sleep(0.2);
    return;
  }

  const sanatoriumID = listRes.json('items.0.id');

  // Use unique date windows per VU/iteration to avoid overlap conflicts.
  const offset = 180 + (__VU * 1000) + (__ITER * 10);
  const checkIn = isoDateUTC(offset);
  const checkOut = isoDateUTC(offset + 7);

  const bookingBody = JSON.stringify({
    sanatorium_id: sanatoriumID,
    check_in: checkIn,
    check_out: checkOut,
    guests: 1,
  });

  const bookingRes = http.post(`${DEAL_BASE}/api/bookings`, bookingBody, {
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    tags: { name: 'create_booking' },
  });

  check(bookingRes, {
    'create booking status is 200/201': (r) => r.status === 200 || r.status === 201,
  });

  sleep(0.2);
}
