Приложения к пояснительной записке

Содержимое приложений размещается **в конце ПЗ** (после списка использованных источников). Листинги — **фрагменты** исходного кода из `src/go-microservices/` и `sql/postgres/`. Полная спецификация атрибутов — в файле `Appendix_B_Specification.md` (эта же папка).

| Приложение | Статус | Назначение |
|:-----------|:-------|:-----------|
| А | справочное | Глоссарий |
| Б | рекомендуемое | Спецификация ключевых таблиц БД (отдельный файл, Б.1–Б.5) |
| В | обязательное | Доменные объекты |
| Г | обязательное | SQL-скрипты |
| Д | обязательное | Слой HTTP-представления (маршруты) |
| Е | обязательное | Слой бизнес-логики |
| Ж | обязательное | Слой доступа к данным |
| И | обязательное | Модульные тесты |
| К | рекомендуемое | Скриншоты тестирования REST API |
| Л | рекомендуемое | Диаграмма Ганта и ресурсы проекта |

---

Приложение А (справочное) Глоссарий

Серверное приложение – это программное обеспечение, которое работает на сервере и предоставляет различные сервисы или функциональность для клиентских приложений. Серверные приложения обычно используются в клиент-серверной архитектуре.

API (Application Programming Interface) – набор правил, протоколов и инструментов, который позволяет различным программным компонентам или системам взаимодействовать друг с другом.

HTTP – протокол передачи гипертекста. Набор правил, по которым данные в интернете передаются между разными источниками, обычно между компьютерами и серверами.

REST API (Representational State Transfer API) – архитектурный стиль для построения web-сервисов, основанный на протоколе HTTP. Характеризуется использованием стандартных HTTP-методов (GET, POST, PUT, DELETE) для взаимодействия с ресурсами, идентифицируемыми по URL.

React – JavaScript-библиотека с открытым исходным кодом для разработки пользовательских интерфейсов, особенно для одностраничных приложений (SPA). Основана на компонентном подходе, что позволяет создавать переиспользуемые и изолированные части интерфейса.

Одностраничное приложение (Single Page Application, SPA) – тип web-приложения, которое загружает одну HTML-страницу и динамически обновляет её содержимое по мере взаимодействия пользователя с приложением, без необходимости полной перезагрузки страницы с сервера.

Слой представления данных – это первый и самый верхний уровень, который присутствует в приложении. Данный уровень представляет собой пользовательский интерфейс (UI), то есть представление содержимого конечному пользователю через интерфейс. В разрабатываемой системе клиентское SPA относится к этому уровню; на стороне микросервисов приём и маршрутизация HTTP-запросов выделены в отдельный слой (приложение Д).

Сервис – это слой бизнес-логики приложения.

Бизнес-логика – это набор правил, необходимых для запуска приложения в соответствии с руководящими принципами, установленными организацией.

Репозиторий – это компонент, отвечающий за взаимодействие с хранилищем данных.

---

Приложение Б (рекомендуемое) Спецификация атрибутов таблиц

См. `Appendix_B_Specification.md` (таблицы Б.1–Б.5 — ключевые таблицы схем `auth`, `deal`, `payment`).

---

Приложение В (обязательное) Доменные объекты

Листинг В.1 – Структура `Booking` (`deal-service/internal/domain/booking.go`)

```go
type Booking struct {
	ID              string     `json:"id"`
	ClientID        string     `json:"client_id"`
	SanatoriumID    string     `json:"sanatorium_id"`
	CheckIn         time.Time  `json:"check_in"`
	CheckOut        time.Time  `json:"check_out"`
	Guests          int        `json:"guests"`
	Status          string     `json:"status"`
	Amount          *float64   `json:"amount,omitempty"`
	Currency        string     `json:"currency,omitempty"`
	PaymentStatus   string     `json:"payment_status,omitempty"`
	InvoiceID       *string    `json:"invoice_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CancelledAt     *time.Time `json:"cancelled_at,omitempty"`
}
```

Листинг В.2 – Структура `Sanatorium` (`deal-service/internal/domain/sanatorium.go`)

```go
type Sanatorium struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	City            string    `json:"city"`
	PricePerNight   float64   `json:"price_per_night"`
	TotalPlaces     int       `json:"total_places"`
	MedicalProfiles []string  `json:"medical_profiles"`
	// ...
}
```

Листинг В.3 – Структуры `Invoice` и `Payment` (`payment-service/internal/domain/payment.go`)

```go
type Invoice struct {
	ID         string    `json:"id"`
	ContractID string    `json:"contract_id,omitempty"`
	BookingID  string    `json:"booking_id,omitempty"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	Status     string    `json:"status"`
}

type Payment struct {
	ID             string    `json:"id"`
	InvoiceID      string    `json:"invoice_id"`
	Amount         float64   `json:"amount"`
	Status         string    `json:"status"`
	IdempotencyKey string    `json:"idempotency_key"`
	PaidAt         time.Time `json:"paid_at"`
}
```

Листинг В.4 – Структура `User` (`auth-service/internal/domain/user.go`)

```go
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	FullName     string    `json:"full_name"`
	Role         string    `json:"role"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}
```

---

Приложение Г (обязательное) SQL-скрипты

Листинг Г.1 – Создание схем и таблицы пользователей (`sql/postgres/00_init_schemas.sql`)

```sql
CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS deal;
CREATE SCHEMA IF NOT EXISTS payment;

CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    full_name TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'client')),
    password_hash TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Листинг Г.2 – Таблица бронирований (`sql/postgres/01_booking_catalog.sql`)

```sql
CREATE TABLE IF NOT EXISTS deal.bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL,
    sanatorium_id UUID NOT NULL REFERENCES deal.sanatoriums(id) ON DELETE RESTRICT,
    check_in DATE NOT NULL,
    check_out DATE NOT NULL,
    guests INT NOT NULL CHECK (guests > 0),
    status TEXT NOT NULL DEFAULT 'confirmed'
        CHECK (status IN ('created', 'confirmed', 'cancelled')),
    CONSTRAINT ck_booking_dates CHECK (check_in < check_out)
);
```

Листинг Г.3 – Связь счёта с бронированием (`sql/postgres/04_booking_payments.sql`)

```sql
ALTER TABLE payment.invoices ADD COLUMN IF NOT EXISTS booking_id UUID NULL;
ALTER TABLE payment.invoices
    ADD CONSTRAINT invoices_reference_check
    CHECK (
        (contract_id IS NOT NULL AND booking_id IS NULL)
        OR (contract_id IS NULL AND booking_id IS NOT NULL)
    );
```

---

Приложение Д (обязательное) Структуры слоя представления данных

Листинг Д.1 – Маршруты `auth-service` (`internal/transport/http/handlers.go`)

```go
func (h *Handler) Router(registry *prometheus.Registry) http.Handler {
	r := chi.NewRouter()
	r.Get("/health", h.ready)
	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/users/register", h.register)
		api.Post("/auth/login", h.login)
	})
	return r
}
```

Листинг Д.2 – Маршруты `deal-service` (`internal/transport/http/handlers.go`)

```go
r.Route("/api", func(api chi.Router) {
	api.Post("/auth/login", h.loginViaAuthService)
	api.Post("/auth/register", h.registerViaAuthService)
	api.Get("/sanatoriums", h.listSanatoriums)
	api.Get("/sanatoriums/{id}", h.getSanatoriumByID)

	api.Group(func(authorized chi.Router) {
		authorized.Use(ClientAuthMiddleware(h.jwtSecret, h.logger))
		authorized.Post("/bookings", h.createBooking)
		authorized.Get("/bookings", h.listBookings)
		authorized.Post("/bookings/{id}/checkout", h.checkoutBooking)
		authorized.Post("/bookings/{id}/pay", h.payBooking)
	})

	api.Route("/admin", func(admin chi.Router) {
		admin.Use(AuthMiddleware(h.jwtSecret, h.logger, "admin"))
		admin.Get("/bookings", h.listBookingsAdmin)
		admin.Post("/bookings/{id}/checkout", h.adminCheckoutBooking)
		admin.Post("/bookings/{id}/pay", h.adminPayBooking)
		admin.Get("/sanatoriums", h.listSanatoriumsAdmin)
		admin.Post("/sanatoriums", h.createSanatoriumAdmin)
	})
})
```

Листинг Д.3 – Маршруты `payment-service` (`internal/transport/http/handlers.go`)

```go
func (h *Handler) Router(registry *prometheus.Registry) http.Handler {
	r := chi.NewRouter()
	r.Post("/internal/invoices", h.createInvoice)
	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/payments", h.processPayment)
		api.Get("/payments/{id}", h.getPayment)
	})
	return r
}
```

---

Приложение Е (обязательное) Слой бизнес-логики

Листинг Е.1 – Метод `BookingService.CreateBooking` (`deal-service/internal/service/booking.go`)

```go
func (s *BookingService) CreateBooking(ctx context.Context, traceID string, in CreateBookingInput) (domain.Booking, error) {
	if strings.TrimSpace(in.ClientID) == "" || strings.TrimSpace(in.SanatoriumID) == "" {
		return domain.Booking{}, fmt.Errorf("client_id and sanatorium_id are required")
	}
	if in.Guests <= 0 {
		return domain.Booking{}, ErrInvalidGuests
	}
	if err := ValidateBookingDateRange(in.CheckIn, in.CheckOut); err != nil {
		return domain.Booking{}, err
	}

	booking, err := s.repo.CreateBooking(ctx, repository.NewBooking{
		ClientID: in.ClientID, SanatoriumID: in.SanatoriumID,
		CheckIn: in.CheckIn, CheckOut: in.CheckOut, Guests: in.Guests,
	})
	if err != nil {
		return domain.Booking{}, err
	}

	s.publishBookingEvent(ctx, traceID, "booking.confirmed", map[string]any{
		"booking_id": booking.ID, "client_id": booking.ClientID,
	})
	return booking, nil
}
```

Листинг Е.2 – Регистрация пользователя (`auth-service/internal/service/auth.go`)

```go
func (s *AuthService) Register(ctx context.Context, email, password, fullName, role string) (domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if !strings.Contains(email, "@") {
		return domain.User{}, ErrInvalidEmail
	}
	if len(password) < 8 {
		return domain.User{}, ErrWeakPassword
	}
	if role != "client" {
		return domain.User{}, ErrInvalidRole
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}
	return s.repo.CreateUser(ctx, email, fullName, role, string(hash))
}
```

Листинг Е.3 – Проведение платежа (`payment-service/internal/service/payment.go`)

```go
func (s *PaymentService) ProcessPayment(ctx context.Context, traceID string, in ProcessPaymentInput) (domain.Payment, bool, error) {
	if in.InvoiceID == "" {
		return domain.Payment{}, false, fmt.Errorf("invoice_id is required")
	}
	if strings.TrimSpace(in.IdempotencyKey) == "" {
		return domain.Payment{}, false, fmt.Errorf("idempotency key is required")
	}
	payment, isDuplicate, err := s.repo.ProcessPayment(ctx, repository.ProcessPaymentInput{
		InvoiceID: in.InvoiceID, Amount: in.Amount,
		IdempotencyKey: in.IdempotencyKey, ExternalRef: in.ExternalRef,
	})
	// публикация payment.completed в NATS при успехе
	return payment, isDuplicate, err
}
```

---

Приложение Ж (обязательное) Слой доступа к данным

Листинг Ж.1 – Создание бронирования в репозитории (`deal-service/internal/repository/booking_catalog_postgres.go`)

```go
func (r *Repository) CreateBooking(ctx context.Context, in NewBooking) (domain.Booking, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return domain.Booking{}, err
	}
	defer tx.Rollback()

	if in.Guests > totalPlaces {
		return domain.Booking{}, ErrGuestsExceedCapacity
	}
	available, err := isSanatoriumAvailable(ctx, tx, in.SanatoriumID, in.CheckIn, in.CheckOut, in.Guests, nil)
	if !available {
		return domain.Booking{}, ErrSanatoriumNotAvailable
	}

	const insertQuery = `
INSERT INTO deal.bookings (id, client_id, sanatorium_id, check_in, check_out, guests, status)
VALUES ($1, $2, $3, $4, $5, $6, 'confirmed')
RETURNING ` + bookingRowColumns
	// ...
	return booking, tx.Commit()
}
```

Листинг Ж.2 – Создание пользователя (`auth-service/internal/repository/postgres.go`)

```go
func (r *UserRepository) CreateUser(ctx context.Context, email, fullName, role, passwordHash string) (domain.User, error) {
	const q = `
INSERT INTO auth.users (email, full_name, role, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING id, email, full_name, role, password_hash, created_at`
	var user domain.User
	err := r.db.QueryRowContext(ctx, q, email, fullName, role, passwordHash).Scan(
		&user.ID, &user.Email, &user.FullName, &user.Role, &user.PasswordHash, &user.CreatedAt,
	)
	return user, err
}
```

---

Приложение И (обязательное) Листинги модульных тестов

Листинг И.1 – Тесты валидации дат (`deal-service/internal/service/booking_test.go`)

```go
func TestValidateBookingDateRange(t *testing.T) {
	checkIn := time.Date(2026, 6, 10, 15, 0, 0, 0, time.UTC)
	checkOut := time.Date(2026, 6, 15, 11, 0, 0, 0, time.UTC)

	if err := ValidateBookingDateRange(checkIn, checkOut); err != nil {
		t.Fatalf("expected valid range, got error: %v", err)
	}
	if err := ValidateBookingDateRange(checkIn, checkIn); err == nil {
		t.Fatalf("expected error for same start/end date")
	}
}
```

Листинг И.2 – Тесты регистрации (`auth-service/internal/service/auth_test.go`)

```go
func TestRegisterAllowsClientOnly(t *testing.T) {
	svc := NewAuthService(stubUserRepo{}, "secret", time.Hour)

	if _, err := svc.Register(context.Background(), "client@example.com", "Pass1234", "Client User", "client"); err != nil {
		t.Fatalf("expected client registration to succeed, got %v", err)
	}
	if _, err := svc.Register(context.Background(), "admin@example.com", "Pass1234", "Admin User", "admin"); err != ErrInvalidRole {
		t.Fatalf("expected ErrInvalidRole for admin registration, got %v", err)
	}
}
```

---

Приложение К (рекомендуемое) Иллюстрационные материалы к тестированию API

Проверка выполнялась через Swagger UI (`/swagger` на `deal-service`), Postman или аналог при работающем стенде Docker Compose. Ниже — перечень ключевых эндпоинтов и места для скриншотов ответов.

Таблица К.0 – Ключевые эндпоинты REST API

| № | Сервис | Метод | Путь | Назначение | Рисунок |
|:--|:-------|:------|:-----|:-----------|:--------|
| 1 | auth-service | POST | `/api/v1/users/register` | Регистрация клиента | К.1 |
| 2 | auth-service | POST | `/api/v1/auth/login` | Выдача JWT | К.2 |
| 3 | deal-service | POST | `/api/auth/login` | Вход через BFF deal-service | К.2 |
| 4 | deal-service | GET | `/api/sanatoriums` | Каталог санаториев | К.3 |
| 5 | deal-service | GET | `/api/sanatoriums/{id}` | Детали и доступность | К.4 |
| 6 | deal-service | POST | `/api/bookings` | Создание бронирования (JWT client) | К.5 |
| 7 | deal-service | POST | `/api/bookings/{id}/checkout` | Выставление счёта | К.6 |
| 8 | deal-service | POST | `/api/bookings/{id}/pay` | Оплата бронирования | К.7 |
| 9 | payment-service | POST | `/api/v1/payments` | Проведение платежа (заголовок `Idempotency-Key`) | К.8 |
| 10 | deal-service | GET | `/api/admin/bookings` | Реестр бронирований (JWT admin) | К.9 |
| 11 | deal-service | POST | `/api/auth/login` | Ошибка 401 при неверном пароле | К.10 |

Примеры тел запросов для скриншотов:

```json
POST /api/v1/users/register
{"email":"client@example.com","password":"Pass1234","full_name":"Иванов И.И.","role":"client"}

POST /api/bookings
{"sanatorium_id":"<uuid>","check_in":"2026-07-10","check_out":"2026-07-15","guests":2}

POST /api/v1/payments
{"invoice_id":"<uuid>","amount":31000.00,"external_ref":"demo-1"}
```

[ЗАГЛУШКА РИСУНКА]

Рисунок К.1 – Успешная регистрация клиента (`POST /api/v1/users/register`, код 201)

[ЗАГЛУШКА РИСУНКА]

Рисунок К.2 – Успешный вход и получение JWT (`POST /api/auth/login`, код 200)

[ЗАГЛУШКА РИСУНКА]

Рисунок К.3 – Получение списка санаториев (`GET /api/sanatoriums`, код 200)

[ЗАГЛУШКА РИСУНКА]

Рисунок К.4 – Детали санатория (`GET /api/sanatoriums/{id}`, код 200)

[ЗАГЛУШКА РИСУНКА]

Рисунок К.5 – Создание бронирования (`POST /api/bookings`, код 201)

[ЗАГЛУШКА РИСУНКА]

Рисунок К.6 – Выставление счёта (`POST /api/bookings/{id}/checkout`, код 200)

[ЗАГЛУШКА РИСУНКА]

Рисунок К.7 – Подтверждение оплаты бронирования (`POST /api/bookings/{id}/pay`, код 200)

[ЗАГЛУШКА РИСУНКА]

Рисунок К.8 – Проведение платежа (`POST /api/v1/payments`, код 200)

[ЗАГЛУШКА РИСУНКА]

Рисунок К.9 – Административный реестр бронирований (`GET /api/admin/bookings`, код 200)

[ЗАГЛУШКА РИСУНКА]

Рисунок К.10 – Ошибка аутентификации (`POST /api/auth/login`, код 401)

---

Приложение Л (рекомендуемое) Графическая часть

Диаграмма Ганта и отчёт о планировании ресурсов разработки ВКР (MS Project, GanttProject или аналог).

[ЗАГЛУШКА РИСУНКА]

Рисунок Л.1 – Диаграмма Ганта этапов разработки

[ЗАГЛУШКА РИСУНКА]

Рисунок Л.2 – Планирование ресурсов проекта
