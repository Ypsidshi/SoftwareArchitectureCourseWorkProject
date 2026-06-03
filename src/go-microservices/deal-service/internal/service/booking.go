package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"coursework/deal-service/internal/domain"
	"coursework/deal-service/internal/repository"
	"coursework/platform-common/pkg/events"
)

var (
	ErrInvalidDateRange = errors.New("invalid booking date range")
	ErrInvalidGuests    = errors.New("guests must be greater than zero")
)

type bookingRepo interface {
	ListSanatoriums(ctx context.Context, filter repository.SanatoriumFilter) ([]domain.Sanatorium, int, error)
	GetSanatoriumByID(ctx context.Context, id string) (domain.Sanatorium, error)
	CheckAvailability(ctx context.Context, sanatoriumID string, checkIn, checkOut time.Time, excludeBookingID *string) (bool, error)
	CreateBooking(ctx context.Context, in repository.NewBooking) (domain.Booking, error)
	UpdateBooking(ctx context.Context, in repository.UpdateBooking) (domain.Booking, error)
	CancelBooking(ctx context.Context, bookingID, clientID string) (domain.Booking, error)
	GetBookingByID(ctx context.Context, bookingID, clientID string) (domain.Booking, error)
	ListBookingsByClient(ctx context.Context, clientID string, page, pageSize int) ([]domain.Booking, int, error)
}

type BookingService struct {
	repo          bookingRepo
	paymentClient any
	publisher     events.Publisher
	serviceName   string
	logger        *slog.Logger
}

type ListSanatoriumsInput struct {
	Page               int
	PageSize           int
	City               string
	ProfileNames       []string
	MaxDistanceToSeaKM *float64
	PriceMin           *float64
	PriceMax           *float64
	CheckIn            *time.Time
	CheckOut           *time.Time
	Sort               string
}

type ListSanatoriumsResult struct {
	Items      []domain.Sanatorium `json:"items"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}

type SanatoriumDetailsResult struct {
	Sanatorium domain.Sanatorium `json:"sanatorium"`
	Available  *bool             `json:"available,omitempty"`
}

type CreateBookingInput struct {
	ClientID     string
	SanatoriumID string
	CheckIn      time.Time
	CheckOut     time.Time
	Guests       int
}

type UpdateBookingInput struct {
	BookingID string
	ClientID  string
	CheckIn   time.Time
	CheckOut  time.Time
	Guests    int
}

type ListBookingsResult struct {
	Items      []domain.Booking `json:"items"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

func NewBookingService(repo bookingRepo, paymentClient any, publisher events.Publisher, serviceName string, logger *slog.Logger) *BookingService {
	return &BookingService{
		repo:          repo,
		paymentClient: paymentClient,
		publisher:     publisher,
		serviceName:   serviceName,
		logger:        logger,
	}
}

func (s *BookingService) ListSanatoriums(ctx context.Context, in ListSanatoriumsInput) (ListSanatoriumsResult, error) {
	page, pageSize := normalizePagination(in.Page, in.PageSize)
	if (in.CheckIn != nil && in.CheckOut == nil) || (in.CheckIn == nil && in.CheckOut != nil) {
		return ListSanatoriumsResult{}, ErrInvalidDateRange
	}
	if in.CheckIn != nil && in.CheckOut != nil {
		if err := ValidateBookingDateRange(*in.CheckIn, *in.CheckOut); err != nil {
			return ListSanatoriumsResult{}, err
		}
	}

	items, total, err := s.repo.ListSanatoriums(ctx, repository.SanatoriumFilter{
		Page:               page,
		PageSize:           pageSize,
		City:               in.City,
		ProfileNames:       in.ProfileNames,
		MaxDistanceToSeaKM: in.MaxDistanceToSeaKM,
		PriceMin:           in.PriceMin,
		PriceMax:           in.PriceMax,
		CheckIn:            in.CheckIn,
		CheckOut:           in.CheckOut,
		Sort:               in.Sort,
	})
	if err != nil {
		return ListSanatoriumsResult{}, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	return ListSanatoriumsResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *BookingService) GetSanatoriumDetails(ctx context.Context, sanatoriumID string, checkIn, checkOut *time.Time) (SanatoriumDetailsResult, error) {
	item, err := s.repo.GetSanatoriumByID(ctx, sanatoriumID)
	if err != nil {
		return SanatoriumDetailsResult{}, err
	}

	var available *bool
	if checkIn != nil && checkOut != nil {
		if err := ValidateBookingDateRange(*checkIn, *checkOut); err != nil {
			return SanatoriumDetailsResult{}, err
		}
		v, err := s.repo.CheckAvailability(ctx, sanatoriumID, *checkIn, *checkOut, nil)
		if err != nil {
			return SanatoriumDetailsResult{}, err
		}
		available = &v
	}

	return SanatoriumDetailsResult{
		Sanatorium: item,
		Available:  available,
	}, nil
}

func (s *BookingService) CreateBooking(ctx context.Context, traceID string, in CreateBookingInput) (domain.Booking, error) {
	if strings.TrimSpace(in.ClientID) == "" || strings.TrimSpace(in.SanatoriumID) == "" {
		return domain.Booking{}, fmt.Errorf("client_id and sanatorium_id are required")
	}
	if in.Guests <= 0 {
		return domain.Booking{}, ErrInvalidGuests
	}
	if err := ValidateBookingDateRange(in.CheckIn, in.CheckOut); err != nil {
		return domain.Booking{}, err
	}

	booking, err := s.repo.CreateBooking(ctx, repository.NewBooking{
		ClientID:     in.ClientID,
		SanatoriumID: in.SanatoriumID,
		CheckIn:      in.CheckIn,
		CheckOut:     in.CheckOut,
		Guests:       in.Guests,
	})
	if err != nil {
		return domain.Booking{}, err
	}

	s.publishBookingEvent(ctx, traceID, "booking.confirmed", map[string]any{
		"booking_id":    booking.ID,
		"client_id":     booking.ClientID,
		"sanatorium_id": booking.SanatoriumID,
		"check_in":      booking.CheckIn,
		"check_out":     booking.CheckOut,
		"guests":        booking.Guests,
		"status":        booking.Status,
	})

	return booking, nil
}

func (s *BookingService) UpdateBooking(ctx context.Context, traceID string, in UpdateBookingInput) (domain.Booking, error) {
	if strings.TrimSpace(in.ClientID) == "" || strings.TrimSpace(in.BookingID) == "" {
		return domain.Booking{}, fmt.Errorf("booking_id and client_id are required")
	}
	if in.Guests <= 0 {
		return domain.Booking{}, ErrInvalidGuests
	}
	if err := ValidateBookingDateRange(in.CheckIn, in.CheckOut); err != nil {
		return domain.Booking{}, err
	}

	booking, err := s.repo.UpdateBooking(ctx, repository.UpdateBooking{
		ID:       in.BookingID,
		ClientID: in.ClientID,
		CheckIn:  in.CheckIn,
		CheckOut: in.CheckOut,
		Guests:   in.Guests,
	})
	if err != nil {
		return domain.Booking{}, err
	}

	s.publishBookingEvent(ctx, traceID, "booking.updated", map[string]any{
		"booking_id":    booking.ID,
		"client_id":     booking.ClientID,
		"sanatorium_id": booking.SanatoriumID,
		"check_in":      booking.CheckIn,
		"check_out":     booking.CheckOut,
		"guests":        booking.Guests,
		"status":        booking.Status,
	})

	return booking, nil
}

func (s *BookingService) CancelBooking(ctx context.Context, traceID string, bookingID, clientID string) (domain.Booking, error) {
	if strings.TrimSpace(bookingID) == "" || strings.TrimSpace(clientID) == "" {
		return domain.Booking{}, fmt.Errorf("booking_id and client_id are required")
	}
	booking, err := s.repo.CancelBooking(ctx, bookingID, clientID)
	if err != nil {
		return domain.Booking{}, err
	}

	s.publishBookingEvent(ctx, traceID, "booking.cancelled", map[string]any{
		"booking_id":    booking.ID,
		"client_id":     booking.ClientID,
		"sanatorium_id": booking.SanatoriumID,
		"status":        booking.Status,
		"cancelled_at":  booking.CancelledAt,
	})
	return booking, nil
}

func (s *BookingService) GetBooking(ctx context.Context, bookingID, clientID string) (domain.Booking, error) {
	return s.repo.GetBookingByID(ctx, bookingID, clientID)
}

func (s *BookingService) ListBookings(ctx context.Context, clientID string, page, pageSize int) (ListBookingsResult, error) {
	page, pageSize = normalizePagination(page, pageSize)
	items, total, err := s.repo.ListBookingsByClient(ctx, clientID, page, pageSize)
	if err != nil {
		return ListBookingsResult{}, err
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	return ListBookingsResult{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *BookingService) publishBookingEvent(ctx context.Context, traceID, eventType string, payload map[string]any) {
	if s.publisher == nil {
		return
	}
	envelope, err := events.NewEnvelope(eventType, s.serviceName, traceID, payload)
	if err != nil {
		s.logger.Error("failed to encode booking event", slog.String("event_type", eventType), slog.String("error", err.Error()))
		return
	}
	if err := s.publisher.Publish(ctx, eventType, envelope); err != nil {
		s.logger.Error("failed to publish booking event", slog.String("event_type", eventType), slog.String("error", err.Error()))
	}
}

func normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func ValidateBookingDateRange(checkIn, checkOut time.Time) error {
	checkIn = toDateUTC(checkIn)
	checkOut = toDateUTC(checkOut)
	if !checkOut.After(checkIn) {
		return ErrInvalidDateRange
	}
	return nil
}

func DatesOverlap(aStart, aEnd, bStart, bEnd time.Time) bool {
	aStart, aEnd = toDateUTC(aStart), toDateUTC(aEnd)
	bStart, bEnd = toDateUTC(bStart), toDateUTC(bEnd)
	return aStart.Before(bEnd) && bStart.Before(aEnd)
}

func toDateUTC(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}

func IsBookingNotFound(err error) bool {
	return errors.Is(err, repository.ErrBookingNotFound)
}

func IsSanatoriumNotFound(err error) bool {
	return errors.Is(err, repository.ErrSanatoriumNotFound)
}

func IsSanatoriumNotAvailable(err error) bool {
	return errors.Is(err, repository.ErrSanatoriumNotAvailable)
}

func IsGuestsExceedCapacity(err error) bool {
	return errors.Is(err, repository.ErrGuestsExceedCapacity)
}
