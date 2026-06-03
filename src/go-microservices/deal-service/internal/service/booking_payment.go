package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"coursework/deal-service/internal/domain"
	"coursework/deal-service/internal/integration/paymentclient"
)

var (
	ErrBookingNotPayable    = errors.New("booking cannot be paid")
	ErrBookingAlreadyPaid   = errors.New("booking is already paid")
	ErrInvoiceNotReady      = errors.New("invoice is not ready for payment")
	ErrPaymentAmountInvalid = errors.New("payment amount does not match booking amount")
)

type bookingPaymentRepo interface {
	GetBookingWithSanatoriumPrice(ctx context.Context, bookingID, clientID string) (domain.Booking, float64, error)
	AttachBookingInvoice(ctx context.Context, bookingID, invoiceID string, amount float64, currency string) error
	MarkBookingInvoiceFailed(ctx context.Context, bookingID, reason string) error
	UpdateBookingPaymentStatus(ctx context.Context, bookingID, status string) error
	GetBookingForClient(ctx context.Context, bookingID, clientID string) (domain.Booking, error)
	ListAllBookings(ctx context.Context, page, pageSize int) ([]domain.Booking, int, error)
}

type bookingPaymentGateway interface {
	CreateInvoice(ctx context.Context, traceID string, req paymentclient.CreateInvoiceRequest) (paymentclient.CreateInvoiceResponse, error)
	ProcessPayment(ctx context.Context, traceID, idempotencyKey string, req paymentclient.ProcessPaymentRequest) (paymentclient.ProcessPaymentResponse, error)
}

func (s *BookingService) CheckoutBooking(ctx context.Context, traceID, bookingID, clientID string) (domain.Booking, error) {
	repo := s.repo.(bookingPaymentRepo)
	booking, pricePerNight, err := repo.GetBookingWithSanatoriumPrice(ctx, bookingID, clientID)
	if err != nil {
		return domain.Booking{}, err
	}
	return s.checkoutBooking(ctx, traceID, bookingID, clientID, booking, pricePerNight, repo)
}

func (s *BookingService) checkoutBooking(
	ctx context.Context,
	traceID, bookingID, _ string,
	booking domain.Booking,
	pricePerNight float64,
	payRepo bookingPaymentRepo,
) (domain.Booking, error) {
	if booking.Status == "cancelled" {
		return domain.Booking{}, ErrBookingNotPayable
	}
	if booking.PaymentStatus == "paid" {
		return domain.Booking{}, ErrBookingAlreadyPaid
	}
	if booking.InvoiceID != nil && booking.PaymentStatus == "invoice_issued" {
		return booking, nil
	}

	amount := calculateBookingAmount(pricePerNight, booking.CheckIn, booking.CheckOut)
	currency := "RUB"
	if booking.Currency != "" {
		currency = booking.Currency
	}

	gateway, ok := s.paymentClient.(bookingPaymentGateway)
	if !ok || gateway == nil {
		return domain.Booking{}, fmt.Errorf("payment client is not configured")
	}

	invoiceResp, err := gateway.CreateInvoice(ctx, traceID, paymentclient.CreateInvoiceRequest{
		BookingID: bookingID,
		Amount:    amount,
		Currency:  currency,
	})
	if err != nil {
		reason := err.Error()
		_ = payRepo.MarkBookingInvoiceFailed(ctx, bookingID, reason)
		return domain.Booking{}, fmt.Errorf("create invoice: %w", err)
	}

	if err := payRepo.AttachBookingInvoice(ctx, bookingID, invoiceResp.InvoiceID, amount, currency); err != nil {
		return domain.Booking{}, err
	}

	if adminRepo, ok := payRepo.(interface {
		GetBookingByIDAdmin(ctx context.Context, bookingID string) (domain.Booking, error)
	}); ok {
		return adminRepo.GetBookingByIDAdmin(ctx, bookingID)
	}
	return payRepo.GetBookingForClient(ctx, bookingID, booking.ClientID)
}

type PayBookingResult struct {
	Booking   domain.Booking `json:"booking"`
	Duplicate bool           `json:"duplicate"`
}

func (s *BookingService) PayBooking(ctx context.Context, traceID, bookingID, clientID, idempotencyKey string) (PayBookingResult, error) {
	repo := s.repo.(bookingPaymentRepo)
	booking, err := repo.GetBookingForClient(ctx, bookingID, clientID)
	if err != nil {
		return PayBookingResult{}, err
	}
	return s.payBooking(ctx, traceID, bookingID, clientID, idempotencyKey, booking, repo)
}

func (s *BookingService) payBooking(
	ctx context.Context,
	traceID, bookingID, clientID, idempotencyKey string,
	booking domain.Booking,
	payRepo bookingPaymentRepo,
) (PayBookingResult, error) {
	if booking.Status == "cancelled" {
		return PayBookingResult{}, ErrBookingNotPayable
	}
	if booking.PaymentStatus == "paid" {
		return PayBookingResult{Booking: booking, Duplicate: false}, nil
	}
	if booking.InvoiceID == nil || booking.Amount == nil {
		return PayBookingResult{}, ErrInvoiceNotReady
	}

	gateway, ok := s.paymentClient.(bookingPaymentGateway)
	if !ok || gateway == nil {
		return PayBookingResult{}, fmt.Errorf("payment client is not configured")
	}

	if strings.TrimSpace(idempotencyKey) == "" {
		idempotencyKey = fmt.Sprintf("booking-%s", bookingID)
	}

	_, err := gateway.ProcessPayment(ctx, traceID, idempotencyKey, paymentclient.ProcessPaymentRequest{
		InvoiceID: *booking.InvoiceID,
		Amount:    *booking.Amount,
	})
	if err != nil {
		return PayBookingResult{}, err
	}

	_ = payRepo.UpdateBookingPaymentStatus(ctx, bookingID, "paid")

	if adminRepo, ok := payRepo.(interface {
		GetBookingByIDAdmin(ctx context.Context, bookingID string) (domain.Booking, error)
	}); ok {
		updated, err := adminRepo.GetBookingByIDAdmin(ctx, bookingID)
		if err != nil {
			return PayBookingResult{}, err
		}
		return PayBookingResult{Booking: updated, Duplicate: false}, nil
	}
	updated, err := payRepo.GetBookingForClient(ctx, bookingID, clientID)
	if err != nil {
		return PayBookingResult{}, err
	}
	return PayBookingResult{Booking: updated, Duplicate: false}, nil
}

func (s *BookingService) HandlePaymentCompleted(ctx context.Context, event PaymentCompleted) error {
	if event.BookingID == "" {
		return nil
	}
	s.logger.Info("payment.completed for booking",
		slog.String("booking_id", event.BookingID),
		slog.String("invoice_id", event.InvoiceID),
		slog.String("payment_id", event.PaymentID))
	return s.repo.(bookingPaymentRepo).UpdateBookingPaymentStatus(ctx, event.BookingID, "paid")
}

func calculateBookingAmount(pricePerNight float64, checkIn, checkOut time.Time) float64 {
	nights := int(math.Ceil(checkOut.Sub(checkIn).Hours() / 24))
	if nights < 1 {
		nights = 1
	}
	return math.Round(pricePerNight*float64(nights)*100) / 100
}
