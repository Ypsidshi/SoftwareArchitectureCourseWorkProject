package domain

import "time"

type Booking struct {
	ID              string     `json:"id"`
	ClientID        string     `json:"client_id"`
	ClientEmail     string     `json:"client_email,omitempty"`
	SanatoriumID    string     `json:"sanatorium_id"`
	SanatoriumName  string     `json:"sanatorium_name,omitempty"`
	CheckIn         time.Time  `json:"check_in"`
	CheckOut        time.Time  `json:"check_out"`
	Guests          int        `json:"guests"`
	Status          string     `json:"status"`
	Amount          *float64   `json:"amount,omitempty"`
	Currency        string     `json:"currency,omitempty"`
	PaymentStatus   string     `json:"payment_status,omitempty"`
	PaymentError    string     `json:"payment_error,omitempty"`
	InvoiceID       *string    `json:"invoice_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CancelledAt     *time.Time `json:"cancelled_at,omitempty"`
}
