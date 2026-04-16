# Coursework Architecture (Go + PostgreSQL)

## Scope
The current implementation contains a client-facing booking flow for a sanatorium domain:
1. Sanatorium catalog with filtering and pagination.
2. Client personal cabinet for booking management.
3. JWT-based access control for booking actions (role: `client`).
4. Async event publishing for booking lifecycle (`booking.confirmed`, `booking.updated`, `booking.cancelled`).
5. Contract creation and payment processing for staff roles (`admin`, `manager`, `accountant`).

## Services
1. `auth-service`
2. `deal-service` (extended with booking/catalog features)
3. `payment-service` (kept for contract/payment flow)

## Booking API (deal-service)
Public:
1. `GET /api/sanatoriums`
2. `GET /api/sanatoriums/{id}`

Authorized client (`Authorization: Bearer <jwt>`):
1. `POST /api/bookings`
2. `GET /api/bookings`
3. `GET /api/bookings/{id}`
4. `PUT /api/bookings/{id}`
5. `DELETE /api/bookings/{id}`

Authorized staff (`Authorization: Bearer <jwt>`, roles: `admin`, `manager`, `accountant`):
1. `POST /api/v1/contracts`
2. `GET /api/v1/contracts/{id}`
3. `PATCH /api/v1/contracts/{id}/status`

Internal:
1. `POST /internal/invoices` on `payment-service` with header `X-Internal-API-Key`

Operational:
1. `GET /health`
2. `GET /health/live`
3. `GET /health/ready`
4. `GET /metrics`
5. `GET /swagger/index.html`

## Database design
New SQL migration:
1. `sql/postgres/01_booking_catalog.sql`

### `auth` schema updates
1. `auth.users.role` now allows: `admin`, `manager`, `accountant`, `client`.

### `deal` schema additions
1. `deal.medical_profiles`
2. `deal.sanatoriums`
3. `deal.sanatorium_medical_profiles` (many-to-many)
4. `deal.bookings`

Booking rules:
1. `check_in < check_out`
2. status lifecycle: `created`, `confirmed`, `cancelled`
3. availability is based on overlapping bookings and cumulative `guests` against `total_places`

## Interaction model
Sync:
1. REST catalog and booking management from client to `deal-service`.
2. Staff creates contracts in `deal-service`.
3. `deal-service` calls `payment-service` to issue invoice via internal API key.

Async:
1. `deal-service` publishes booking lifecycle events to NATS.
2. `payment-service` publishes `payment.completed` to NATS.
3. `deal-service` subscribes to `payment.completed` and marks contract payment as `paid`.

## Swagger
`deal-service` has endpoint annotations for `swaggo/swag`.
Generated artifacts:
1. `src/go-microservices/deal-service/docs/docs.go`
2. `src/go-microservices/deal-service/docs/swagger.json`
3. `src/go-microservices/deal-service/docs/swagger.yaml`

## Testing status
Implemented:
1. Unit tests for booking date validation and overlap logic:
`deal-service/internal/service/booking_test.go`
2. Unit tests for booking capacity calculation:
`deal-service/internal/repository/booking_catalog_postgres_test.go`
3. Unit tests for exact invoice amount checks:
`payment-service/internal/repository/postgres_test.go`
4. Unit tests for public registration restrictions:
`auth-service/internal/service/auth_test.go`

Build verification:
1. `go test ./...`
2. `go build ./...`

## Next diploma-ready steps
1. Add DB migrations runner (`golang-migrate`) in service startup.
2. Add integration tests (PostgreSQL + NATS with Testcontainers).
3. Add outbox/inbox for guaranteed event delivery.
4. Add admin panel APIs for catalog management.
5. Replace shared JWT secret validation with common typed auth package or asymmetric signing.
6. Add bootstrap strategy for privileged staff accounts instead of public self-registration.
