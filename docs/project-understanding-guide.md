# Project Understanding Guide

Этот документ помогает быстро понять, из чего состоит проект, какая часть является основной, какие технологии используются и как движутся данные между частями системы.

## 1. Что это за проект

Сейчас в репозитории есть **две линии развития**:

- **Основная и рабочая**: `src/go-microservices`
- **Ранняя/черновая .NET-заготовка**: `src/services` и `src/gateway`

Для понимания текущего проекта нужно смотреть прежде всего на Go-реализацию.

## 2. Карта всего репозитория

```mermaid
flowchart TD
    repo[SoftwareArchitectureCourseWorkProject]

    repo --> docs[docs]
    repo --> sql[sql]
    repo --> src[src]
    repo --> testing[Testing]
    repo --> textFiles[text_files]
    repo --> extracted[._docx_extract]

    docs --> docsMain[coursework and audit docs]
    docs --> docsDiagrams[architecture diagrams]

    sql --> sqlPg[sql/postgres]
    sql --> sqlMs[sql/00_schema_full_mssql.sql]

    src --> goMain[src/go-microservices]
    src --> dotnetServices[src/services]
    src --> dotnetGateway[src/gateway]

    goMain --> auth[auth-service]
    goMain --> deal[deal-service]
    goMain --> payment[payment-service]
    goMain --> common[platform-common]
    goMain --> compose[docker-compose and go.work]

    testing --> reports[reports, k6, metrics, examples]
    textFiles --> drafts[drafts, reports, pict]
    extracted --> extractedDocs[temporary extracted docx contents]
```

## 3. Какие части проекта реально важны

### Актуальная система

- `src/go-microservices/auth-service`
- `src/go-microservices/deal-service`
- `src/go-microservices/payment-service`
- `src/go-microservices/platform-common`
- `sql/postgres`
- `docs`

### Неосновная часть

- `src/services`
- `src/gateway`

Эти каталоги сейчас больше похожи на след от раннего варианта архитектуры на `.NET`: там есть шаблонные `Program.cs`, `weatherforecast`, `Class1.cs`, но почти нет предметной логики.

### Поддерживающие материалы

- `Testing`
- `text_files`
- часть generated и учебных документов

Они важны для отчётов и учебного процесса, но не определяют, как работает основная система.

## 4. Технологический стек

```mermaid
flowchart LR
    subgraph frontendSide [Clients]
        client[Client]
        staff[Staff]
    end

    subgraph backend [Backend]
        go[Go 1.26 workspace]
        chi[Chi Router]
        jwt[JWT]
        pgx[pgx / database/sql]
        nats[NATS]
        prom[Prometheus metrics]
        swag[Swagger for deal-service]
    end

    subgraph infra [Infrastructure]
        pg[PostgreSQL]
        docker[Docker Compose]
    end

    client --> go
    staff --> go

    go --> chi
    go --> jwt
    go --> pgx
    go --> nats
    go --> prom
    go --> swag
    go --> pg
    go --> docker
```

## 5. Основная runtime-архитектура

```mermaid
flowchart LR
    subgraph clients [Users]
        client[Client]
        staff[Staff]
    end

    subgraph services [Go Microservices]
        authService[auth-service]
        dealService[deal-service]
        paymentService[payment-service]
    end

    subgraph commonLayer [Shared Layer]
        platformCommon[platform-common]
    end

    subgraph asyncBus [Async]
        nats[NATS]
    end

    subgraph storage [Storage]
        postgres[(PostgreSQL)]
    end

    client -->|"register/login"| authService
    client -->|"catalog + bookings"| dealService
    staff -->|"contracts"| dealService
    staff -->|"payments"| paymentService

    dealService -->|"login proxy"| authService
    dealService -->|"issue invoice"| paymentService

    paymentService -->|"publish payment.completed"| nats
    dealService -->|"publish booking.*"| nats
    nats -->|"deliver payment.completed"| dealService

    authService --> postgres
    dealService --> postgres
    paymentService --> postgres

    authService --> platformCommon
    dealService --> platformCommon
    paymentService --> platformCommon
```

## 6. За что отвечает каждый сервис

```mermaid
flowchart TD
    authService[auth-service]
    dealService[deal-service]
    paymentService[payment-service]
    common[platform-common]

    authService --> authResp[Registration, login, JWT issuing]
    dealService --> dealResp[Catalog, bookings, contracts, payment status sync]
    paymentService --> payResp[Invoice creation, payment processing, payment event]
    common --> commonResp[Logging, recovery, trace-id, metrics, NATS envelope]
```

### `auth-service`

- регистрация пользователя;
- логин;
- выпуск JWT;
- хранение пользователей в схеме `auth`.

### `deal-service`

- каталог санаториев;
- фильтрация и просмотр карточек;
- создание, изменение и отмена бронирований;
- создание договоров;
- обновление статуса оплаты договора после события из `payment-service`.

### `payment-service`

- создание инвойса для договора;
- обработка платежа;
- публикация события `payment.completed`;
- хранение инвойсов и платежей в схеме `payment`.

### `platform-common`

- middleware;
- trace id;
- JSON helper;
- логирование;
- Prometheus metrics;
- общие обёртки вокруг NATS.

## 7. Внутренняя структура Go-сервисов

Почти каждый сервис устроен одинаково:

```mermaid
flowchart LR
    cmd[cmd/service/main.go] --> transport[transport/http or transport/events]
    transport --> service[internal/service]
    service --> repository[internal/repository]
    repository --> db[(PostgreSQL)]

    service --> integration[integration clients]
    integration --> external[other services]
```

Это значит:

- `cmd` собирает приложение и зависимости;
- `transport` принимает HTTP-запросы или события;
- `service` содержит бизнес-логику;
- `repository` работает с БД;
- `integration` вызывает другие сервисы.

## 8. Как проходит логин

```mermaid
sequenceDiagram
    participant Client
    participant Deal as deal-service
    participant Auth as auth-service
    participant DB as PostgreSQL

    Client->>Deal: POST /api/auth/login
    Deal->>Auth: POST /api/v1/auth/login
    Auth->>DB: find user by email
    Auth->>Auth: check password and create JWT
    Auth-->>Deal: access_token + user
    Deal-->>Client: access_token + user
```

Особенность: пользователь логинится через `deal-service`, но сам JWT выпускает `auth-service`.

## 9. Как проходит бронирование

```mermaid
sequenceDiagram
    participant Client
    participant Deal as deal-service
    participant DB as PostgreSQL
    participant Bus as NATS

    Client->>Deal: POST /api/bookings
    Deal->>Deal: validate JWT and role=client
    Deal->>DB: check dates and capacity
    Deal->>DB: create booking
    Deal->>Bus: publish booking.confirmed
    Deal-->>Client: booking response
```

После аудита логика доступности теперь учитывает не просто факт пересечения броней, а суммарное число гостей против `total_places`.

## 10. Как проходит договор и оплата

```mermaid
sequenceDiagram
    participant Staff
    participant Deal as deal-service
    participant Payment as payment-service
    participant DB as PostgreSQL
    participant Bus as NATS

    Staff->>Deal: POST /api/v1/contracts
    Deal->>Deal: validate JWT and staff role
    Deal->>DB: create contract
    Deal->>Payment: POST /internal/invoices
    Payment->>Payment: check X-Internal-API-Key
    Payment->>DB: create invoice
    Payment-->>Deal: invoice_id
    Deal->>DB: attach invoice to contract
    Deal-->>Staff: contract with invoice info

    Staff->>Payment: POST /api/v1/payments
    Payment->>DB: validate invoice and exact amount
    Payment->>DB: store payment and mark invoice paid
    Payment->>Bus: publish payment.completed
    Bus-->>Deal: payment.completed
    Deal->>DB: update contract payment_status=paid
```

## 11. Как устроена база данных

```mermaid
erDiagram
    AUTH_USERS {
        uuid id
        string email
        string full_name
        string role
    }

    DEAL_SANATORIUMS {
        uuid id
        string name
        string city
        numeric price_per_night
        int total_places
    }

    DEAL_BOOKINGS {
        uuid id
        uuid client_id
        uuid sanatorium_id
        date check_in
        date check_out
        int guests
        string status
    }

    DEAL_CONTRACTS {
        uuid id
        uuid resident_id
        uuid manager_id
        numeric amount
        string payment_status
        uuid invoice_id
    }

    PAYMENT_INVOICES {
        uuid id
        uuid contract_id
        numeric amount
        string currency
        string status
    }

    PAYMENT_PAYMENTS {
        uuid id
        uuid invoice_id
        numeric amount
        string status
        string idempotency_key
    }

    DEAL_SANATORIUMS ||--o{ DEAL_BOOKINGS : has
    AUTH_USERS ||--o{ DEAL_BOOKINGS : creates
    DEAL_CONTRACTS ||--|| PAYMENT_INVOICES : generates
    PAYMENT_INVOICES ||--o{ PAYMENT_PAYMENTS : receives
```

Ключевая идея: одна PostgreSQL используется как общее физическое хранилище, но логически она разбита на схемы `auth`, `deal`, `payment`.

## 12. Что с .NET-частью

```mermaid
flowchart LR
    dotnet[.NET branch]
    go[Go branch]

    dotnet --> dotnetState[Template-level skeleton]
    go --> goState[Current working implementation]
```

### Что это означает

- `.NET`-ветка важна как исторический след или альтернативный трек;
- но **не она определяет, как работает проект сейчас**;
- для защиты и дальнейшего развития лучше объяснять именно Go-архитектуру.

## 13. Сильные стороны проекта

- уже есть рабочее разделение на микросервисы;
- есть Docker Compose для локального запуска;
- есть health, metrics, structured logging, trace-id;
- сервисы разделены по ролям и ответственности;
- есть event-driven часть через NATS;
- уже виден переход от курсовой к дипломной архитектуре.

## 14. Ограничения, которые важно понимать

- события работают по MVP-схеме, без надёжной доставки уровня production;
- JWT пока основан на shared secret и локальной проверке в сервисах;
- часть репозитория засорена учебными и документными артефактами;
- покрытие тестами ещё небольшое;
- `.NET`-ветка может путать, если не проговорить, что она не основная.

## 15. В каком порядке изучать проект

1. `docs/project-understanding-guide.md`
2. `docs/project-readiness-audit.md`
3. `src/go-microservices/docker-compose.yml`
4. `src/go-microservices/README.md`
5. `src/go-microservices/auth-service`
6. `src/go-microservices/deal-service`
7. `src/go-microservices/payment-service`
8. `src/go-microservices/platform-common`
9. `sql/postgres`

## 16. Полезные связанные файлы

- `docs/project-readiness-audit.md`
- `docs/project-self-learning-roadmap.md`
- `docs/coursework-go.md`
- `docs/diagrams/go-runtime-overview.mmd`
- `docs/diagrams/booking-payment-sequence.mmd`
- `docs/diagrams/domain-data-model.mmd`
