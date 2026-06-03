package service

import "context"

type PaymentEventsRouter struct {
	Deal    *DealService
	Booking *BookingService
}

func (r *PaymentEventsRouter) HandlePaymentCompleted(ctx context.Context, event PaymentCompleted) error {
	if event.BookingID != "" && r.Booking != nil {
		return r.Booking.HandlePaymentCompleted(ctx, event)
	}
	if event.ContractID != "" && r.Deal != nil {
		return r.Deal.HandlePaymentCompleted(ctx, event)
	}
	return nil
}
