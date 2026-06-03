package service

import (
	"context"

	"coursework/deal-service/internal/domain"
	"coursework/deal-service/internal/repository"
)

type AdminBookingsFilter struct {
	Status        string
	PaymentStatus string
	City          string
	SanatoriumID  string
}

type adminBookingRepo interface {
	bookingPaymentRepo
	GetBookingByIDAdmin(ctx context.Context, bookingID string) (domain.Booking, error)
	GetBookingWithSanatoriumPriceAdmin(ctx context.Context, bookingID string) (domain.Booking, float64, error)
	CancelBookingAdmin(ctx context.Context, bookingID string) (domain.Booking, error)
	ListAllBookingsFiltered(ctx context.Context, page, pageSize int, filter repository.AdminBookingsFilter) ([]domain.Booking, int, error)
}

func (s *BookingService) ListBookingsAdmin(ctx context.Context, page, pageSize int, filter AdminBookingsFilter) (ListBookingsResult, error) {
	page, pageSize = normalizePagination(page, pageSize)
	repo := s.repo.(adminBookingRepo)
	items, total, err := repo.ListAllBookingsFiltered(ctx, page, pageSize, repository.AdminBookingsFilter{
		Status:        filter.Status,
		PaymentStatus: filter.PaymentStatus,
		City:          filter.City,
		SanatoriumID:  filter.SanatoriumID,
	})
	if err != nil {
		return ListBookingsResult{}, err
	}
	totalPages := 0
	if pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	return ListBookingsResult{
		Items: items, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages,
	}, nil
}

func (s *BookingService) AdminCancelBooking(ctx context.Context, traceID, bookingID string) (domain.Booking, error) {
	repo := s.repo.(adminBookingRepo)
	booking, err := repo.CancelBookingAdmin(ctx, bookingID)
	if err != nil {
		return domain.Booking{}, err
	}
	s.publishBookingEvent(ctx, traceID, "booking.cancelled", map[string]any{
		"booking_id": booking.ID, "client_id": booking.ClientID, "sanatorium_id": booking.SanatoriumID,
	})
	return booking, nil
}

func (s *BookingService) AdminCheckoutBooking(ctx context.Context, traceID, bookingID string) (domain.Booking, error) {
	repo := s.repo.(adminBookingRepo)
	booking, pricePerNight, err := repo.GetBookingWithSanatoriumPriceAdmin(ctx, bookingID)
	if err != nil {
		return domain.Booking{}, err
	}
	return s.checkoutBooking(ctx, traceID, bookingID, booking.ClientID, booking, pricePerNight, repo)
}

func (s *BookingService) AdminPayBooking(ctx context.Context, traceID, bookingID, idempotencyKey string) (PayBookingResult, error) {
	repo := s.repo.(adminBookingRepo)
	booking, err := repo.GetBookingByIDAdmin(ctx, bookingID)
	if err != nil {
		return PayBookingResult{}, err
	}
	return s.payBooking(ctx, traceID, bookingID, booking.ClientID, idempotencyKey, booking, repo)
}
