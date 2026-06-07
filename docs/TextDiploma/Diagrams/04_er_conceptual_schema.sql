-- Минимальная схема для рис. 4 (концептуальная модель, клиентский контур).
-- Для reverse engineering в Visual Paradigm / pgModeler / DBeaver:
--   1) поднять Postgres (docker compose в src/go-microservices);
--   2) либо выполнить этот файл в отдельной БД concept_er;
--   3) либо в VP выбрать только перечисленные таблицы из sanatorium.
-- Служебные: auth.refresh_tokens, deal.contracts — не включать в рис. 4.

CREATE SCHEMA IF NOT EXISTS concept;

CREATE TABLE concept.users (
    id          UUID PRIMARY KEY,
    email       TEXT NOT NULL UNIQUE,
    full_name   TEXT NOT NULL,
    role        TEXT NOT NULL
);

CREATE TABLE concept.sanatoriums (
    id              UUID PRIMARY KEY,
    name            TEXT NOT NULL,
    city            TEXT NOT NULL,
    price_per_night NUMERIC(12, 2) NOT NULL,
    total_places    INT NOT NULL
);

CREATE TABLE concept.medical_profiles (
    id          UUID PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    description TEXT
);

CREATE TABLE concept.sanatorium_medical_profiles (
    sanatorium_id UUID NOT NULL REFERENCES concept.sanatoriums (id),
    profile_id    UUID NOT NULL REFERENCES concept.medical_profiles (id),
    PRIMARY KEY (sanatorium_id, profile_id)
);

CREATE TABLE concept.bookings (
    id             UUID PRIMARY KEY,
    client_id      UUID NOT NULL REFERENCES concept.users (id),
    sanatorium_id  UUID NOT NULL REFERENCES concept.sanatoriums (id),
    check_in       DATE NOT NULL,
    check_out      DATE NOT NULL,
    guests         INT NOT NULL,
    status         TEXT NOT NULL
);

CREATE TABLE concept.invoices (
    id         UUID PRIMARY KEY,
    booking_id UUID NOT NULL UNIQUE REFERENCES concept.bookings (id),
    amount     NUMERIC(12, 2) NOT NULL,
    currency   CHAR(3) NOT NULL,
    status     TEXT NOT NULL
);

CREATE TABLE concept.payments (
    id               UUID PRIMARY KEY,
    invoice_id       UUID NOT NULL REFERENCES concept.invoices (id),
    amount           NUMERIC(12, 2) NOT NULL,
    status           TEXT NOT NULL,
    idempotency_key  TEXT NOT NULL UNIQUE
);
