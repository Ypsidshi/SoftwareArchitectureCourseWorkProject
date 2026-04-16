CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE SCHEMA IF NOT EXISTS auth;
CREATE SCHEMA IF NOT EXISTS deal;
CREATE SCHEMA IF NOT EXISTS payment;

CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    full_name TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'manager', 'accountant')),
    password_hash TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auth_users_role ON auth.users(role);

CREATE TABLE IF NOT EXISTS auth.refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_auth_refresh_tokens_user_id ON auth.refresh_tokens(user_id);

CREATE TABLE IF NOT EXISTS deal.contracts (
    id UUID PRIMARY KEY,
    resident_id UUID NOT NULL,
    room_id UUID NOT NULL,
    manager_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    amount NUMERIC(12, 2) NOT NULL CHECK (amount > 0),
    currency CHAR(3) NOT NULL DEFAULT 'RUB',
    status TEXT NOT NULL DEFAULT 'created' CHECK (status IN ('created', 'confirmed', 'cancelled', 'completed')),
    payment_status TEXT NOT NULL DEFAULT 'invoice_requested' CHECK (
        payment_status IN ('invoice_requested', 'invoice_issued', 'invoice_failed', 'paid')
    ),
    payment_error TEXT NOT NULL DEFAULT '',
    invoice_id UUID NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_deal_contract_dates CHECK (start_date <= end_date)
);

CREATE INDEX IF NOT EXISTS idx_deal_contracts_resident_id ON deal.contracts(resident_id);
CREATE INDEX IF NOT EXISTS idx_deal_contracts_room_id ON deal.contracts(room_id);
CREATE INDEX IF NOT EXISTS idx_deal_contracts_manager_id ON deal.contracts(manager_id);
CREATE INDEX IF NOT EXISTS idx_deal_contracts_status ON deal.contracts(status);
CREATE INDEX IF NOT EXISTS idx_deal_contracts_payment_status ON deal.contracts(payment_status);
CREATE UNIQUE INDEX IF NOT EXISTS uq_deal_contracts_invoice_id ON deal.contracts(invoice_id) WHERE invoice_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS payment.invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id UUID NOT NULL UNIQUE,
    amount NUMERIC(12, 2) NOT NULL CHECK (amount > 0),
    currency CHAR(3) NOT NULL DEFAULT 'RUB',
    status TEXT NOT NULL DEFAULT 'issued' CHECK (status IN ('issued', 'paid', 'overdue', 'cancelled')),
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_invoices_status ON payment.invoices(status);
CREATE INDEX IF NOT EXISTS idx_payment_invoices_contract_id ON payment.invoices(contract_id);

CREATE TABLE IF NOT EXISTS payment.payments (
    id UUID PRIMARY KEY,
    invoice_id UUID NOT NULL REFERENCES payment.invoices(id) ON DELETE CASCADE,
    amount NUMERIC(12, 2) NOT NULL CHECK (amount > 0),
    status TEXT NOT NULL CHECK (status IN ('completed', 'failed', 'refunded')),
    idempotency_key TEXT NOT NULL UNIQUE,
    external_ref TEXT NULL,
    paid_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_payments_invoice_id ON payment.payments(invoice_id);
CREATE INDEX IF NOT EXISTS idx_payment_payments_paid_at ON payment.payments(paid_at);
