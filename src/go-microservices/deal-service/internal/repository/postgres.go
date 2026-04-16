package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"coursework/deal-service/internal/domain"
)

var ErrNotFound = errors.New("contract not found")

type Repository struct {
	db *sql.DB
}

type NewContract struct {
	ID         string
	ResidentID string
	RoomID     string
	ManagerID  string
	StartDate  time.Time
	EndDate    time.Time
	Amount     float64
	Currency   string
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *Repository) CreateContract(ctx context.Context, c NewContract) (domain.Contract, error) {
	const q = `
INSERT INTO deal.contracts (
	id, resident_id, room_id, manager_id, start_date, end_date, amount, currency, status, payment_status
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'created', 'invoice_requested')
RETURNING id, resident_id, room_id, manager_id, start_date, end_date, amount, currency,
          status, payment_status, payment_error, invoice_id, created_at, updated_at`

	var item domain.Contract
	err := r.db.QueryRowContext(
		ctx, q,
		c.ID, c.ResidentID, c.RoomID, c.ManagerID, c.StartDate, c.EndDate, c.Amount, c.Currency,
	).Scan(
		&item.ID,
		&item.ResidentID,
		&item.RoomID,
		&item.ManagerID,
		&item.StartDate,
		&item.EndDate,
		&item.Amount,
		&item.Currency,
		&item.Status,
		&item.PaymentStatus,
		&item.PaymentError,
		&item.InvoiceID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return domain.Contract{}, err
	}
	return item, nil
}

func (r *Repository) AttachInvoice(ctx context.Context, contractID, invoiceID string) error {
	const q = `
UPDATE deal.contracts
SET invoice_id = $2, payment_status = 'invoice_issued', payment_error = '', updated_at = NOW()
WHERE id = $1`

	res, err := r.db.ExecContext(ctx, q, contractID, invoiceID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) MarkInvoiceFailed(ctx context.Context, contractID, reason string) error {
	const q = `
UPDATE deal.contracts
SET payment_status = 'invoice_failed', payment_error = $2, updated_at = NOW()
WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, contractID, reason)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) UpdateStatus(ctx context.Context, contractID, status string) error {
	const q = `
UPDATE deal.contracts
SET status = $2, updated_at = NOW()
WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, contractID, status)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) UpdatePaymentStatus(ctx context.Context, contractID, paymentStatus string) error {
	const q = `
UPDATE deal.contracts
SET payment_status = $2, payment_error = '', updated_at = NOW()
WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, contractID, paymentStatus)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (domain.Contract, error) {
	const q = `
SELECT id, resident_id, room_id, manager_id, start_date, end_date, amount, currency,
       status, payment_status, payment_error, invoice_id, created_at, updated_at
FROM deal.contracts
WHERE id = $1`

	var item domain.Contract
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&item.ID,
		&item.ResidentID,
		&item.RoomID,
		&item.ManagerID,
		&item.StartDate,
		&item.EndDate,
		&item.Amount,
		&item.Currency,
		&item.Status,
		&item.PaymentStatus,
		&item.PaymentError,
		&item.InvoiceID,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Contract{}, ErrNotFound
		}
		return domain.Contract{}, fmt.Errorf("query contract: %w", err)
	}

	return item, nil
}
