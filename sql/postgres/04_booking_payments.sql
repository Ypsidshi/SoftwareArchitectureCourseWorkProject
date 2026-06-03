-- Booking payments and demo admin account.

ALTER TABLE deal.bookings
    ADD COLUMN IF NOT EXISTS amount NUMERIC(12, 2) NULL,
    ADD COLUMN IF NOT EXISTS currency CHAR(3) NOT NULL DEFAULT 'RUB',
    ADD COLUMN IF NOT EXISTS payment_status TEXT NOT NULL DEFAULT 'unpaid',
    ADD COLUMN IF NOT EXISTS payment_error TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS invoice_id UUID NULL;

ALTER TABLE deal.bookings DROP CONSTRAINT IF EXISTS bookings_payment_status_check;
ALTER TABLE deal.bookings
    ADD CONSTRAINT bookings_payment_status_check
    CHECK (payment_status IN ('unpaid', 'invoice_issued', 'invoice_failed', 'paid'));

CREATE UNIQUE INDEX IF NOT EXISTS uq_deal_bookings_invoice_id
    ON deal.bookings (invoice_id) WHERE invoice_id IS NOT NULL;

ALTER TABLE payment.invoices
    ADD COLUMN IF NOT EXISTS booking_id UUID NULL;

ALTER TABLE payment.invoices ALTER COLUMN contract_id DROP NOT NULL;

ALTER TABLE payment.invoices DROP CONSTRAINT IF EXISTS invoices_reference_check;
ALTER TABLE payment.invoices
    ADD CONSTRAINT invoices_reference_check
    CHECK (
        (contract_id IS NOT NULL AND booking_id IS NULL)
        OR (contract_id IS NULL AND booking_id IS NOT NULL)
    );

CREATE UNIQUE INDEX IF NOT EXISTS uq_payment_invoices_booking_id
    ON payment.invoices (booking_id) WHERE booking_id IS NOT NULL;

-- Demo admin seed moved to 05_seed_demo_admin.sql (password hash must match Admin1234!)
