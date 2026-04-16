# Go Microservices Coursework

This folder contains the Go implementation of the coursework platform.

## Services
1. `auth-service`
2. `deal-service` (contracts + client booking/catalog API)
3. `payment-service`
4. `platform-common` (shared middleware, metrics, events)

## Run with Docker
```bash
cd src/go-microservices
docker compose up --build
```

After startup:
1. `auth-service`: `http://localhost:8081`
2. `deal-service`: `http://localhost:8082`
3. `payment-service`: `http://localhost:8083`
4. `nats`: `nats://localhost:4222`

## DB initialization
Executed automatically by PostgreSQL container:
1. `sql/postgres/00_init_schemas.sql`
2. `sql/postgres/01_booking_catalog.sql`

## Booking API quick check
1. Register client:
```bash
curl -X POST http://localhost:8081/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{"email":"client@example.com","password":"Pass1234","full_name":"Client User","role":"client"}'
```

2. Login and get JWT:
```bash
curl -X POST http://localhost:8082/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"client@example.com","password":"Pass1234"}'
```

3. Catalog list:
```bash
curl "http://localhost:8082/api/sanatoriums?page=1&page_size=10&city=Sochi"
```

4. Create booking:
```bash
curl -X POST http://localhost:8082/api/bookings \
  -H "Authorization: Bearer <JWT_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{"sanatorium_id":"<sanatorium_uuid>","check_in":"2026-07-10","check_out":"2026-07-20","guests":1}'
```

5. Swagger UI:
`http://localhost:8082/swagger/index.html`

## Contracts and payments
- `POST /api/v1/contracts`, `GET /api/v1/contracts/{id}` and `PATCH /api/v1/contracts/{id}/status` in `deal-service` now require JWT with one of roles: `admin`, `manager`, `accountant`.
- Public self-registration is intentionally limited to `role=client`. Privileged staff accounts should be provisioned separately for demo or seeded directly in the database.
- `payment-service` accepts `/internal/invoices` only with `X-Internal-API-Key`. The shared key is configured through `INTERNAL_API_KEY` in Docker Compose.
- Payment processing now requires the payment amount to exactly match the invoice amount.

## Health and metrics
Each service provides:
1. `/health`
2. `/health/live`
3. `/health/ready`
4. `/metrics`
