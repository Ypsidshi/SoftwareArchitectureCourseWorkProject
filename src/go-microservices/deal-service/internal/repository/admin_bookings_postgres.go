package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"coursework/deal-service/internal/domain"
)

type AdminBookingsFilter struct {
	Status         string
	PaymentStatus  string
	City           string
	SanatoriumID   string
}

func (r *Repository) GetBookingByIDAdmin(ctx context.Context, bookingID string) (domain.Booking, error) {
	const query = `
SELECT ` + bookingSelectColumns + `
FROM deal.bookings b
WHERE b.id = $1`

	var booking domain.Booking
	row := r.db.QueryRowContext(ctx, query, bookingID)
	if err := scanBooking(row, &booking); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Booking{}, ErrBookingNotFound
		}
		return domain.Booking{}, err
	}
	return booking, nil
}

func (r *Repository) GetBookingWithSanatoriumPriceAdmin(ctx context.Context, bookingID string) (domain.Booking, float64, error) {
	const query = `
SELECT ` + bookingSelectColumns + `, s.price_per_night
FROM deal.bookings b
JOIN deal.sanatoriums s ON s.id = b.sanatorium_id
WHERE b.id = $1`

	var booking domain.Booking
	var pricePerNight float64
	row := r.db.QueryRowContext(ctx, query, bookingID)
	var amount sql.NullFloat64
	var currency sql.NullString
	var paymentStatus sql.NullString
	var paymentError sql.NullString
	var invoiceID sql.NullString
	err := row.Scan(
		&booking.ID, &booking.ClientID, &booking.SanatoriumID,
		&booking.CheckIn, &booking.CheckOut, &booking.Guests, &booking.Status,
		&booking.CreatedAt, &booking.UpdatedAt, &booking.CancelledAt,
		&amount, &currency, &paymentStatus, &paymentError, &invoiceID,
		&pricePerNight,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Booking{}, 0, ErrBookingNotFound
		}
		return domain.Booking{}, 0, err
	}
	applyBookingPaymentNulls(&booking, amount, currency, paymentStatus, paymentError, invoiceID)
	return booking, pricePerNight, nil
}

func (r *Repository) CancelBookingAdmin(ctx context.Context, bookingID string) (domain.Booking, error) {
	const query = `
UPDATE deal.bookings
SET status = 'cancelled', cancelled_at = NOW(), updated_at = NOW()
WHERE id = $1 AND status <> 'cancelled'
RETURNING ` + bookingRowColumns

	var booking domain.Booking
	err := scanBooking(r.db.QueryRowContext(ctx, query, bookingID), &booking)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Booking{}, ErrBookingNotFound
		}
		return domain.Booking{}, err
	}
	return booking, nil
}

func (r *Repository) ListAllBookingsFiltered(ctx context.Context, page, pageSize int, filter AdminBookingsFilter) ([]domain.Booking, int, error) {
	where, args := buildAdminBookingsWhere(filter)

	countQuery := `SELECT COUNT(*) FROM deal.bookings b JOIN deal.sanatoriums s ON s.id = b.sanatorium_id LEFT JOIN auth.users u ON u.id = b.client_id WHERE ` + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limitParam := len(args) + 1
	offsetParam := len(args) + 2
	args = append(args, pageSize, (page-1)*pageSize)
	listQuery := `
SELECT
	b.id, b.client_id, b.sanatorium_id, s.name,
	b.check_in, b.check_out, b.guests, b.status,
	b.created_at, b.updated_at, b.cancelled_at,
	b.amount, b.currency, b.payment_status, b.payment_error, b.invoice_id,
	COALESCE(u.email, '') AS client_email
FROM deal.bookings b
JOIN deal.sanatoriums s ON s.id = b.sanatorium_id
LEFT JOIN auth.users u ON u.id = b.client_id
WHERE ` + where + `
ORDER BY b.created_at DESC
LIMIT $` + strconv.Itoa(limitParam) + ` OFFSET $` + strconv.Itoa(offsetParam)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]domain.Booking, 0, pageSize)
	for rows.Next() {
		var booking domain.Booking
		if err := scanAdminBookingEmail(rows, &booking); err != nil {
			return nil, 0, err
		}
		items = append(items, booking)
	}
	return items, total, rows.Err()
}

func buildAdminBookingsWhere(filter AdminBookingsFilter) (string, []any) {
	conditions := []string{"1=1"}
	args := make([]any, 0, 4)
	add := func(cond string, val any) {
		args = append(args, val)
		conditions = append(conditions, fmt.Sprintf(cond, len(args)))
	}
	if s := strings.TrimSpace(filter.Status); s != "" {
		add("b.status = $%d", s)
	}
	if s := strings.TrimSpace(filter.PaymentStatus); s != "" {
		add("b.payment_status = $%d", s)
	}
	if s := strings.TrimSpace(filter.City); s != "" {
		add("LOWER(s.city) = LOWER($%d)", s)
	}
	if s := strings.TrimSpace(filter.SanatoriumID); s != "" {
		add("b.sanatorium_id = $%d", s)
	}
	return strings.Join(conditions, " AND "), args
}

func applyBookingPaymentNulls(booking *domain.Booking, amount sql.NullFloat64, currency, paymentStatus, paymentError, invoiceID sql.NullString) {
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
}

func scanAdminBookingEmail(sc interface {
	Scan(dest ...any) error
}, booking *domain.Booking) error {
	var amount sql.NullFloat64
	var currency sql.NullString
	var paymentStatus sql.NullString
	var paymentError sql.NullString
	var invoiceID sql.NullString
	err := sc.Scan(
		&booking.ID, &booking.ClientID, &booking.SanatoriumID, &booking.SanatoriumName,
		&booking.CheckIn, &booking.CheckOut, &booking.Guests, &booking.Status,
		&booking.CreatedAt, &booking.UpdatedAt, &booking.CancelledAt,
		&amount, &currency, &paymentStatus, &paymentError, &invoiceID,
		&booking.ClientEmail,
	)
	if err != nil {
		return err
	}
	applyBookingPaymentNulls(booking, amount, currency, paymentStatus, paymentError, invoiceID)
	return nil
}
