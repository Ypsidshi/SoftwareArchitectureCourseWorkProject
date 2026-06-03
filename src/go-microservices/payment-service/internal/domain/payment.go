package domain

import "time"

type Invoice struct {
	ID         string    `json:"id"`
	ContractID string    `json:"contract_id,omitempty"`
	BookingID  string    `json:"booking_id,omitempty"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	Status     string    `json:"status"`
	IssuedAt   time.Time `json:"issued_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Payment struct {
	ID             string    `json:"id"`
	InvoiceID      string    `json:"invoice_id"`
	ContractID     string    `json:"contract_id,omitempty"`
	BookingID      string    `json:"booking_id,omitempty"`
	Amount         float64   `json:"amount"`
	Status         string    `json:"status"`
	IdempotencyKey string    `json:"idempotency_key"`
	ExternalRef    string    `json:"external_ref,omitempty"`
	PaidAt         time.Time `json:"paid_at"`
	CreatedAt      time.Time `json:"created_at"`
}
