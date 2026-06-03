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

## UI and API gateway
The SPA talks to **deal-service** for catalog, bookings, admin APIs, and **auth proxy** (`POST /api/auth/login`, `POST /api/auth/register`). Payment is internal (deal → payment-service); the browser does not call payment-service directly.

## Demo accounts (UI login at http://localhost:3000/login)

| Role   | Email                    | Password    |
|--------|--------------------------|-------------|
| Admin  | `admin@sanatorium.local` | `Admin1234!` |
| Client | register at `/register`  | (your choice) |

Admin is created by SQL init (`sql/postgres/05_seed_demo_admin.sql`). If login fails on an **old** database volume, run:

```bash
docker exec -i coursework-postgres psql -U postgres -d sanatorium < ../../sql/postgres/05_seed_demo_admin.sql
```

After startup:
1. `auth-service`: `http://localhost:8081`
2. `deal-service`: `http://localhost:8082`
3. `payment-service`: `http://localhost:8083`
4. `nats`: `nats://localhost:4222`

## DB initialization
Executed automatically by PostgreSQL container (`sql/postgres/` in repo root):
1. `00_init_schemas.sql` — schemas, roles
2. `01_booking_catalog.sql` — catalog, bookings
3. `02_update_sanatorium_images.sql`, `03_upsert_extra_sanatoriums.sql` — seed data
4. `04_booking_payments.sql` — payment columns on bookings
5. `05_seed_demo_admin.sql` — demo admin account
6. `06_real_sanatorium_images.sql` — real photo URLs (apply to existing DB without reset)
7. `08_dedupe_sanatoriums.sql` — remove duplicate catalog rows (e.g. double «Рассвет»)
8. `09_expand_medical_profiles.sql` — расширить справочник медпрофилей и привязки к санаториям

## Booking API quick check
1. Register client (via deal-service auth proxy):
```bash
curl -X POST http://localhost:8082/api/auth/register \
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

Regenerate OpenAPI (from `deal-service/`):
```bash
go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/deal-service/main.go -o docs --parseDependency --parseInternal
```

## Roles
- **client** — registration, catalog, bookings, pay for a booking (`/bookings` in UI).
- **admin** — panel at `/admin` (bookings, contracts, sanatoriums). See **Demo accounts** above.

### Admin API (`/api/admin`, JWT role `admin`)
- **Bookings:** `GET /bookings` (filters: status, payment_status, city, sanatorium_id), `DELETE /bookings/{id}`, `POST /bookings/{id}/checkout`, `POST /bookings/{id}/pay`.
- **Sanatoriums:** `GET/POST /sanatoriums`, `PUT /sanatoriums/{id}`, `DELETE /sanatoriums/{id}` (delete blocked if active confirmed bookings exist); medical profiles from catalog (`GET /api/medical-profiles`).

## Booking payment (client)
- `POST /api/bookings/{id}/checkout` — price = nights × `price_per_night`, creates invoice.
- `POST /api/bookings/{id}/pay` — pays invoice (optional `Idempotency-Key` header).

## Notes
- Public registration is limited to `role=client`.
- `payment-service` `/internal/invoices` requires `X-Internal-API-Key` (`INTERNAL_API_KEY` in Compose).
- Payment amount must exactly match the invoice amount.

## Health and metrics
Each service provides:
1. `/health`
2. `/health/live`
3. `/health/ready`
4. `/metrics`
