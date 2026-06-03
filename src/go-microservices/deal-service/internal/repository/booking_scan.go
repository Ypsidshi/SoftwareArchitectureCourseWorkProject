package repository

import (
	"database/sql"

	"coursework/deal-service/internal/domain"
)

const bookingRowColumns = `
id, client_id, sanatorium_id, check_in, check_out, guests, status,
created_at, updated_at, cancelled_at,
amount, currency, payment_status, payment_error, invoice_id`

const bookingSelectColumns = `
b.id, b.client_id, b.sanatorium_id, b.check_in, b.check_out, b.guests, b.status,
b.created_at, b.updated_at, b.cancelled_at,
b.amount, b.currency, b.payment_status, b.payment_error, b.invoice_id`

func scanBooking(sc interface {
	Scan(dest ...any) error
}, booking *domain.Booking) error {
	var amount sql.NullFloat64
	var currency sql.NullString
	var paymentStatus sql.NullString
	var paymentError sql.NullString
	var invoiceID sql.NullString

	err := sc.Scan(
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
	)
	if err != nil {
		return err
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
	return nil
}

func scanAdminBooking(sc interface {
	Scan(dest ...any) error
}, booking *domain.Booking) error {
	var amount sql.NullFloat64
	var currency sql.NullString
	var paymentStatus sql.NullString
	var paymentError sql.NullString
	var invoiceID sql.NullString

	err := sc.Scan(
		&booking.ID,
		&booking.ClientID,
		&booking.SanatoriumID,
		&booking.SanatoriumName,
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
	)
	if err != nil {
		return err
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
	return nil
}
