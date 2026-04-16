package domain

import "time"

type Booking struct {
	ID           string     `json:"id"`
	ClientID     string     `json:"client_id"`
	SanatoriumID string     `json:"sanatorium_id"`
	CheckIn      time.Time  `json:"check_in"`
	CheckOut     time.Time  `json:"check_out"`
	Guests       int        `json:"guests"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	CancelledAt  *time.Time `json:"cancelled_at,omitempty"`
}
