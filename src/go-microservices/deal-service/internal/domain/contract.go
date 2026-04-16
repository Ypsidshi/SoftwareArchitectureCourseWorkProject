package domain

import "time"

type Contract struct {
	ID            string    `json:"id"`
	ResidentID    string    `json:"resident_id"`
	RoomID        string    `json:"room_id"`
	ManagerID     string    `json:"manager_id"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	PaymentStatus string    `json:"payment_status"`
	PaymentError  string    `json:"payment_error,omitempty"`
	InvoiceID     *string   `json:"invoice_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
