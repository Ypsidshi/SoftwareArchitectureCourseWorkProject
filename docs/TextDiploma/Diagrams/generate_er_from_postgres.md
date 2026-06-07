# Рисунок 4 — как получить ER в стиле примера (VP / IDEF1X)

PlantUML для концептуальной модели даёт упрощённую картинку без атрибутов.  
Образец из диплома — **логико-концептуальный уровень в Visual Paradigm, нотация IDEF1X**: сущности с полями, ключи, «воронья лапа».

## Что даёт автоматика из PostgreSQL

| Источник | Результат | Подходит для рис. 4 |
|----------|-----------|---------------------|
| Reverse из живой БД `sanatorium` | Все физические таблицы (`refresh_tokens`, `contracts`, …) | Нужно **отфильтровать** лишнее |
| Файл `04_er_conceptual_schema.sql` | Только 6+1 сущностей клиентского контура | **Да** |
| Файл `04_er_conceptual.dbml` | Та же структура, без Postgres | **Да** (dbdiagram.io) |

Полностью «концептуальную» картинку без технических полей Postgres **не** построит сам: после reverse в VP скрывают `created_at`, `updated_at`, переименовывают таблицы на русский.

---

## Вариант A — Visual Paradigm 17.x (как в образцах диплома)

1. Запустить БД:
   ```powershell
   cd src/go-microservices
   docker compose up -d postgres
   ```
2. VP: **Tools → Database → Reverse Database to Model…**
3. Подключение:
   - Host: `localhost`, Port: `5432`
   - Database: `sanatorium`
   - User / Password: `postgres` / `postgres`
4. Схемы: `auth`, `deal`, `payment` (или только `concept` после выполнения `04_er_conceptual_schema.sql`).
5. Для рис. 4 оставить таблицы:
   - `auth.users`
   - `deal.sanatoriums`, `deal.medical_profiles`, `deal.sanatorium_medical_profiles`, `deal.bookings`
   - `payment.invoices`, `payment.payments`
6. **Не включать:** `refresh_tokens`, `contracts` (служебный контур).
7. Нотация диаграммы: **IDEF1X** (свойства диаграммы).
8. Переименовать сущности по таблице 1 ПЗ (Пользователь, Санаторий, …).
9. Удалить или скрыть служебные атрибуты (`created_at`, `updated_at`, `password_hash` на концептуальном уровне по желанию методру).
10. **File → Export Diagram → PNG** — вставить в Word вместо заглушки рис. 4.

Альтернатива без фильтрации: выполнить `04_er_conceptual_schema.sql` в новой БД и сделать reverse только схемы `concept`.

---

## Вариант B — dbdiagram.io (быстро, без VP)

1. Открыть https://dbdiagram.io/d
2. **Import** → файл `04_er_conceptual.dbml`
3. При необходимости подправить расположение блоков.
4. **Export → PNG**
5. Для диплома: при необходимости перенести в VP и переключить на IDEF1X / ч/б.

---

## Вариант C — DBeaver / pgModeler

- **DBeaver:** подключиться к `sanatorium` → ER Diagram → выбрать нужные таблицы → Export PNG.
- **pgModeler:** File → Import → SQL file → `04_er_conceptual_schema.sql` → Export PNG.

Стиль ближе к Crow's Foot, чем к IDEF1X VP; для защиты часто достаточно, если комиссия не требует строго VP.

---

## Соответствие таблицам 1–2 ПЗ

| Сущность ПЗ | Таблица |
|-------------|---------|
| Пользователь | users |
| Санаторий | sanatoriums |
| Медицинский профиль | medical_profiles |
| M:N | sanatorium_medical_profiles |
| Бронирование | bookings |
| Счёт на оплату | invoices |
| Платёж | payments |

Связь M:N в ER-инструментах обычно показывается через связующую таблицу (как «Избранное» в вашем примере) — это нормально для IDEF1X.

---

# Рисунок 6 — физическая модель (фрагмент)

Текст §2.2.3: **фрагмент схем `deal` и `payment`**, типы PostgreSQL, PK/FK, индексы.  
Полная спецификация полей — **приложение Б** (таблицы Б.1–Б.5). На рисунке — не вся логическая модель (рис. 5).

## Чем физическая отличается от логической (рис. 5)

| Рис. 5 (логическая) | Рис. 6 (физическая) |
|---------------------|---------------------|
| 9 таблиц, все схемы | **4 таблицы** (фрагмент) |
| Имена и связи | **Точные типы** (`UUID`, `NUMERIC(12,2)`, `TIMESTAMPTZ`, …) |
| Пунктирные межсхемные связи | FK там, где есть в БД; остальное — note |
| `refresh_tokens`, `contracts`, каталог M:N | **Не включать** на рис. 6 |

## Таблицы для рис. 6

| ✓ | Таблица | Зачем |
|---|---------|-------|
| ✓ | `deal.sanatoriums` | FK из `bookings`; индексы каталога (§2.2.3) |
| ✓ | `deal.bookings` | Ядро сценария |
| ✓ | `payment.invoices` | Checkout |
| ✓ | `payment.payments` | Оплата, UK `idempotency_key` |

**Не брать:** `auth.*`, `medical_profiles`, `sanatorium_medical_profiles`, `contracts`, `refresh_tokens`.

## Связи на рис. 6

| Связь | Кратность | Линия | Поля |
|-------|-----------|-------|------|
| sanatoriums → bookings | 1:N | **сплошная** | `bookings.sanatorium_id` → FK |
| bookings ↔ invoices | 1:1 | **пунктир** | `bookings.invoice_id`, `invoices.booking_id` (межсхемно, без FK) |
| invoices → payments | 1:N | **сплошная** | `payments.invoice_id` → FK |

**Примечание** у `invoices`: CHECK — ровно одно из `contract_id`, `booking_id` (на фрагменте `contract_id` можно оставить полем без таблицы `contracts`).

## Индексы и ограничения (упомянуть на рисунке или в note)

**deal.sanatoriums:** `idx_sanatoriums_city`, `idx_sanatoriums_price`, `idx_sanatoriums_distance`, `uq_deal_sanatoriums_name_city`.

**deal.bookings:** `idx_bookings_client`, `idx_bookings_sanatorium`, `idx_bookings_dates`, `uq_deal_bookings_invoice_id`; CHECK `check_in < check_out`.

**payment.invoices:** `uq_payment_invoices_booking_id`; CHECK XOR с `contract_id`.

**payment.payments:** `idx_payment_payments_invoice_id`; UNIQUE `idempotency_key`.

Расширение `btree_gist` упомянуто в §2.2.3; EXCLUDE-ограничение в SQL пока не создано — на диаграмме не рисовать.

## Шаги в Visual Paradigm

1. **Новая** ERD: *Add Diagram → Entity Relationship Diagram* → имя `Physical fragment`.
2. Перетащить 4 таблицы из модели (или reverse только `deal.sanatoriums`, `deal.bookings`, `payment.invoices`, `payment.payments`).
3. Нотация **IDEF1X**; уровень **Physical** (типы колонок из Postgres).
4. Связи — по таблице выше; убрать артефакты VP (`userid`, `contractsid`, …).
5. Ч/б экспорт → `docs/TextDiploma/Diagrams/06_er_physical_fragment.png`.

## Рис. 13 (раздел 3)

Тот же или ещё более узкий фрагмент (`bookings` + `invoices`) — «реализованная структура»; можно скрин pgAdmin/DBeaver или копию рис. 6.
