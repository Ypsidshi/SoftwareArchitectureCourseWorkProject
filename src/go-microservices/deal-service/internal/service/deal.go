package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"coursework/deal-service/internal/domain"
	"coursework/deal-service/internal/integration/paymentclient"
	"coursework/deal-service/internal/repository"
	"github.com/google/uuid"
)

var (
	ErrInvalidStatus  = errors.New("invalid contract status")
	ErrInvoiceFailure = errors.New("invoice creation failed")
)

type dealRepo interface {
	Ping(ctx context.Context) error
	CreateContract(ctx context.Context, c repository.NewContract) (domain.Contract, error)
	AttachInvoice(ctx context.Context, contractID, invoiceID string) error
	MarkInvoiceFailed(ctx context.Context, contractID, reason string) error
	UpdateStatus(ctx context.Context, contractID, status string) error
	UpdatePaymentStatus(ctx context.Context, contractID, paymentStatus string) error
	GetByID(ctx context.Context, id string) (domain.Contract, error)
}

type paymentGateway interface {
	CreateInvoice(ctx context.Context, traceID string, req paymentclient.CreateInvoiceRequest) (paymentclient.CreateInvoiceResponse, error)
}

type DealService struct {
	repo          dealRepo
	paymentClient paymentGateway
	logger        *slog.Logger
}

type CreateContractInput struct {
	ResidentID string
	RoomID     string
	ManagerID  string
	StartDate  time.Time
	EndDate    time.Time
	Amount     float64
	Currency   string
}

type PaymentCompleted struct {
	ContractID string  `json:"contract_id"`
	InvoiceID  string  `json:"invoice_id"`
	PaymentID  string  `json:"payment_id"`
	Amount     float64 `json:"amount"`
}

func NewDealService(repo dealRepo, paymentClient paymentGateway, logger *slog.Logger) *DealService {
	return &DealService{
		repo:          repo,
		paymentClient: paymentClient,
		logger:        logger,
	}
}

func (s *DealService) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s *DealService) CreateContract(ctx context.Context, traceID string, in CreateContractInput) (domain.Contract, error) {
	if in.Amount <= 0 {
		return domain.Contract{}, fmt.Errorf("amount must be positive")
	}
	if !in.EndDate.After(in.StartDate) {
		return domain.Contract{}, fmt.Errorf("end_date must be after start_date")
	}
	if in.ResidentID == "" || in.RoomID == "" || in.ManagerID == "" {
		return domain.Contract{}, fmt.Errorf("resident_id, room_id and manager_id are required")
	}

	contractID := uuid.NewString()
	entity, err := s.repo.CreateContract(ctx, repository.NewContract{
		ID:         contractID,
		ResidentID: in.ResidentID,
		RoomID:     in.RoomID,
		ManagerID:  in.ManagerID,
		StartDate:  in.StartDate,
		EndDate:    in.EndDate,
		Amount:     in.Amount,
		Currency:   strings.ToUpper(strings.TrimSpace(in.Currency)),
	})
	if err != nil {
		return domain.Contract{}, err
	}

	invoiceResp, err := s.paymentClient.CreateInvoice(ctx, traceID, paymentclient.CreateInvoiceRequest{
		ContractID: entity.ID,
		Amount:     entity.Amount,
		Currency:   entity.Currency,
	})
	if err != nil {
		reason := err.Error()
		if len(reason) > 300 {
			reason = reason[:300]
		}
		_ = s.repo.MarkInvoiceFailed(ctx, entity.ID, reason)
		entity.PaymentStatus = "invoice_failed"
		entity.PaymentError = reason
		return entity, fmt.Errorf("%w: %v", ErrInvoiceFailure, err)
	}

	if err := s.repo.AttachInvoice(ctx, entity.ID, invoiceResp.InvoiceID); err != nil {
		return entity, err
	}
	entity.InvoiceID = &invoiceResp.InvoiceID
	entity.PaymentStatus = "invoice_issued"
	entity.PaymentError = ""
	return entity, nil
}

func (s *DealService) GetContract(ctx context.Context, id string) (domain.Contract, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *DealService) UpdateStatus(ctx context.Context, id, status string) error {
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "created", "confirmed", "cancelled", "completed":
	default:
		return ErrInvalidStatus
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

func (s *DealService) HandlePaymentCompleted(ctx context.Context, event PaymentCompleted) error {
	if event.ContractID == "" {
		return fmt.Errorf("contract_id is required")
	}
	s.logger.Info("payment.completed received",
		slog.String("contract_id", event.ContractID),
		slog.String("invoice_id", event.InvoiceID),
		slog.String("payment_id", event.PaymentID))
	return s.repo.UpdatePaymentStatus(ctx, event.ContractID, "paid")
}

func IsNotFound(err error) bool {
	return errors.Is(err, repository.ErrNotFound)
}
