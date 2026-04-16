package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"coursework/payment-service/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrInvoiceNotFound = errors.New("invoice not found")
	ErrPaymentNotFound = errors.New("payment not found")
	ErrInvalidAmount   = errors.New("invalid payment amount")
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *Repository) CreateInvoice(ctx context.Context, contractID string, amount float64, currency string) (domain.Invoice, error) {
	const q = `
INSERT INTO payment.invoices (contract_id, amount, currency, status)
VALUES ($1, $2, $3, 'issued')
RETURNING id, contract_id, amount, currency, status, issued_at, updated_at`

	var invoice domain.Invoice
	err := r.db.QueryRowContext(ctx, q, contractID, amount, currency).Scan(
		&invoice.ID,
		&invoice.ContractID,
		&invoice.Amount,
		&invoice.Currency,
		&invoice.Status,
		&invoice.IssuedAt,
		&invoice.UpdatedAt,
	)
	if err != nil {
		return domain.Invoice{}, err
	}
	return invoice, nil
}

type ProcessPaymentInput struct {
	InvoiceID      string
	Amount         float64
	IdempotencyKey string
	ExternalRef    string
}

func (r *Repository) ProcessPayment(ctx context.Context, in ProcessPaymentInput) (domain.Payment, bool, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return domain.Payment{}, false, err
	}
	defer tx.Rollback()

	if in.IdempotencyKey != "" {
		existing, err := queryPaymentByIdempotency(ctx, tx, in.IdempotencyKey)
		if err == nil {
			if err := tx.Commit(); err != nil {
				return domain.Payment{}, false, err
			}
			return existing, true, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return domain.Payment{}, false, err
		}
	}

	const lockInvoiceQuery = `
SELECT id, contract_id, amount, status
FROM payment.invoices
WHERE id = $1
FOR UPDATE`

	var invoiceID, contractID, status string
	var invoiceAmount float64
	err = tx.QueryRowContext(ctx, lockInvoiceQuery, in.InvoiceID).Scan(&invoiceID, &contractID, &invoiceAmount, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Payment{}, false, ErrInvoiceNotFound
		}
		return domain.Payment{}, false, err
	}

	if !isExactInvoiceAmount(invoiceAmount, in.Amount) {
		return domain.Payment{}, false, ErrInvalidAmount
	}

	if status == "paid" {
		// Defensive path for cases without idempotency key.
		const q = `
SELECT p.id, p.invoice_id, i.contract_id, p.amount, p.status, p.idempotency_key, p.external_ref, p.paid_at, p.created_at
FROM payment.payments p
JOIN payment.invoices i ON i.id = p.invoice_id
WHERE p.invoice_id = $1
ORDER BY p.created_at DESC
LIMIT 1`
		var payment domain.Payment
		if err := tx.QueryRowContext(ctx, q, in.InvoiceID).Scan(
			&payment.ID, &payment.InvoiceID, &payment.ContractID, &payment.Amount, &payment.Status,
			&payment.IdempotencyKey, &payment.ExternalRef, &payment.PaidAt, &payment.CreatedAt,
		); err != nil {
			return domain.Payment{}, false, err
		}
		if err := tx.Commit(); err != nil {
			return domain.Payment{}, false, err
		}
		return payment, true, nil
	}

	const insertPayment = `
INSERT INTO payment.payments (id, invoice_id, amount, status, idempotency_key, external_ref, paid_at)
VALUES ($1, $2, $3, 'completed', $4, $5, NOW())
RETURNING id, invoice_id, amount, status, idempotency_key, external_ref, paid_at, created_at`

	var payment domain.Payment
	paymentID := uuid.NewString()
	err = tx.QueryRowContext(ctx, insertPayment, paymentID, in.InvoiceID, in.Amount, in.IdempotencyKey, in.ExternalRef).Scan(
		&payment.ID,
		&payment.InvoiceID,
		&payment.Amount,
		&payment.Status,
		&payment.IdempotencyKey,
		&payment.ExternalRef,
		&payment.PaidAt,
		&payment.CreatedAt,
	)
	if err != nil {
		return domain.Payment{}, false, err
	}

	const updateInvoice = `
UPDATE payment.invoices
SET status = 'paid', updated_at = NOW()
WHERE id = $1`
	if _, err := tx.ExecContext(ctx, updateInvoice, in.InvoiceID); err != nil {
		return domain.Payment{}, false, err
	}

	payment.ContractID = contractID

	if err := tx.Commit(); err != nil {
		return domain.Payment{}, false, err
	}
	return payment, false, nil
}

func queryPaymentByIdempotency(ctx context.Context, tx *sql.Tx, idempotencyKey string) (domain.Payment, error) {
	const q = `
SELECT p.id, p.invoice_id, i.contract_id, p.amount, p.status, p.idempotency_key, p.external_ref, p.paid_at, p.created_at
FROM payment.payments p
JOIN payment.invoices i ON i.id = p.invoice_id
WHERE p.idempotency_key = $1`

	var payment domain.Payment
	err := tx.QueryRowContext(ctx, q, idempotencyKey).Scan(
		&payment.ID,
		&payment.InvoiceID,
		&payment.ContractID,
		&payment.Amount,
		&payment.Status,
		&payment.IdempotencyKey,
		&payment.ExternalRef,
		&payment.PaidAt,
		&payment.CreatedAt,
	)
	return payment, err
}

func (r *Repository) GetPaymentByID(ctx context.Context, id string) (domain.Payment, error) {
	const q = `
SELECT p.id, p.invoice_id, i.contract_id, p.amount, p.status, p.idempotency_key, p.external_ref, p.paid_at, p.created_at
FROM payment.payments p
JOIN payment.invoices i ON i.id = p.invoice_id
WHERE p.id = $1`

	var payment domain.Payment
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&payment.ID,
		&payment.InvoiceID,
		&payment.ContractID,
		&payment.Amount,
		&payment.Status,
		&payment.IdempotencyKey,
		&payment.ExternalRef,
		&payment.PaidAt,
		&payment.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Payment{}, ErrPaymentNotFound
		}
		return domain.Payment{}, fmt.Errorf("query payment: %w", err)
	}
	return payment, nil
}

func (r *Repository) MarkInvoiceOverdue(ctx context.Context, olderThan time.Time) (int64, error) {
	const q = `
UPDATE payment.invoices
SET status = 'overdue', updated_at = NOW()
WHERE status = 'issued' AND issued_at < $1`
	res, err := r.db.ExecContext(ctx, q, olderThan)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func isExactInvoiceAmount(invoiceAmount, paymentAmount float64) bool {
	return paymentAmount > 0 && paymentAmount == invoiceAmount
}
