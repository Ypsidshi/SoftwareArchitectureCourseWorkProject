package repository

import (
	"context"
	"database/sql"
	"errors"

	"coursework/deal-service/internal/domain"
)

func (r *Repository) GetBookingForClient(ctx context.Context, bookingID, clientID string) (domain.Booking, error) {
	const query = `
SELECT ` + bookingSelectColumns + `
FROM deal.bookings b
WHERE b.id = $1 AND b.client_id = $2`

	var booking domain.Booking
	row := r.db.QueryRowContext(ctx, query, bookingID, clientID)
	if err := scanBooking(row, &booking); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Booking{}, ErrBookingNotFound
		}
		return domain.Booking{}, err
	}
	return booking, nil
}

func (r *Repository) GetBookingWithSanatoriumPrice(ctx context.Context, bookingID, clientID string) (domain.Booking, float64, error) {
	const query = `
SELECT ` + bookingSelectColumns + `, s.price_per_night
FROM deal.bookings b
JOIN deal.sanatoriums s ON s.id = b.sanatorium_id
WHERE b.id = $1 AND b.client_id = $2`

	var booking domain.Booking
	var pricePerNight float64
	row := r.db.QueryRowContext(ctx, query, bookingID, clientID)
	var amount sql.NullFloat64
	var currency sql.NullString
	var paymentStatus sql.NullString
	var paymentError sql.NullString
	var invoiceID sql.NullString
	err := row.Scan(
		&booking.ID,
		&booking.ClientID,
		&booking.SanatoriumID,
		&booking.CheckIn,
		&booking.CheckOut,
		&booking.Guests,
		&booking.Status,
		&booking.CreatedAt,
		&booking.UpdatedAt,
		&booking.CancelledAt,
		&amount,
		&currency,
		&paymentStatus,
		&paymentError,
		&invoiceID,
		&pricePerNight,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Booking{}, 0, ErrBookingNotFound
		}
		return domain.Booking{}, 0, err
	}
	if amount.Valid {
		v := amount.Float64
		booking.Amount = &v
	}
	if currency.Valid {
		booking.Currency = currency.String
	}
	if paymentStatus.Valid {
		booking.PaymentStatus = paymentStatus.String
	}
	if paymentError.Valid {
		booking.PaymentError = paymentError.String
	}
	if invoiceID.Valid {
		v := invoiceID.String
		booking.InvoiceID = &v
	}
	return booking, pricePerNight, nil
}

func (r *Repository) AttachBookingInvoice(ctx context.Context, bookingID, invoiceID string, amount float64, currency string) error {
	const q = `
UPDATE deal.bookings
SET invoice_id = $2,
    amount = $3,
    currency = $4,
    payment_status = 'invoice_issued',
    payment_error = '',
    updated_at = NOW()
WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, bookingID, invoiceID, amount, currency)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrBookingNotFound
	}
	return nil
}

func (r *Repository) MarkBookingInvoiceFailed(ctx context.Context, bookingID, reason string) error {
	if len(reason) > 300 {
		reason = reason[:300]
	}
	const q = `
UPDATE deal.bookings
SET payment_status = 'invoice_failed',
    payment_error = $2,
    updated_at = NOW()
WHERE id = $1`
	_, err := r.db.ExecContext(ctx, q, bookingID, reason)
	return err
}

func (r *Repository) UpdateBookingPaymentStatus(ctx context.Context, bookingID, status string) error {
	const q = `
UPDATE deal.bookings
SET payment_status = $2,
    payment_error = '',
    updated_at = NOW()
WHERE id = $1`
	res, err := r.db.ExecContext(ctx, q, bookingID, status)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrBookingNotFound
	}
	return nil
}

func (r *Repository) ListAllBookings(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error) {
	const countQuery = `SELECT COUNT(*) FROM deal.bookings`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	const listQuery = `
SELECT
	b.id, b.client_id, b.sanatorium_id, s.name,
	b.check_in, b.check_out, b.guests, b.status,
	b.created_at, b.updated_at, b.cancelled_at,
	b.amount, b.currency, b.payment_status, b.payment_error, b.invoice_id
FROM deal.bookings b
JOIN deal.sanatoriums s ON s.id = b.sanatorium_id
ORDER BY b.created_at DESC
LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, listQuery, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.Booking, 0, pageSize)
	for rows.Next() {
		var booking domain.Booking
		if err := scanAdminBooking(rows, &booking); err != nil {
			return nil, 0, err
		}
		items = append(items, booking)
	}
	return items, total, rows.Err()
}
