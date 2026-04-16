package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"coursework/payment-service/internal/domain"
	"coursework/payment-service/internal/repository"
	"coursework/platform-common/pkg/events"
)

type paymentRepo interface {
	Ping(ctx context.Context) error
	CreateInvoice(ctx context.Context, contractID string, amount float64, currency string) (domain.Invoice, error)
	ProcessPayment(ctx context.Context, in repository.ProcessPaymentInput) (domain.Payment, bool, error)
	GetPaymentByID(ctx context.Context, id string) (domain.Payment, error)
}

type PaymentService struct {
	repo        paymentRepo
	publisher   events.Publisher
	serviceName string
	logger      *slog.Logger
}

type CreateInvoiceInput struct {
	ContractID string
	Amount     float64
	Currency   string
}

type ProcessPaymentInput struct {
	InvoiceID      string
	Amount         float64
	IdempotencyKey string
	ExternalRef    string
}

func NewPaymentService(repo paymentRepo, publisher events.Publisher, serviceName string, logger *slog.Logger) *PaymentService {
	return &PaymentService{
		repo:        repo,
		publisher:   publisher,
		serviceName: serviceName,
		logger:      logger,
	}
}

func (s *PaymentService) Ping(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s *PaymentService) CreateInvoice(ctx context.Context, in CreateInvoiceInput) (domain.Invoice, error) {
	if in.ContractID == "" {
		return domain.Invoice{}, fmt.Errorf("contract_id is required")
	}
	if in.Amount <= 0 {
		return domain.Invoice{}, fmt.Errorf("amount must be positive")
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = "RUB"
	}
	return s.repo.CreateInvoice(ctx, in.ContractID, in.Amount, currency)
}

func (s *PaymentService) ProcessPayment(ctx context.Context, traceID string, in ProcessPaymentInput) (domain.Payment, bool, error) {
	if in.InvoiceID == "" {
		return domain.Payment{}, false, fmt.Errorf("invoice_id is required")
	}
	if strings.TrimSpace(in.IdempotencyKey) == "" {
		return domain.Payment{}, false, fmt.Errorf("idempotency key is required")
	}

	payment, isDuplicate, err := s.repo.ProcessPayment(ctx, repository.ProcessPaymentInput{
		InvoiceID:      in.InvoiceID,
		Amount:         in.Amount,
		IdempotencyKey: in.IdempotencyKey,
		ExternalRef:    in.ExternalRef,
	})
	if err != nil {
		return domain.Payment{}, false, err
	}

	if isDuplicate || s.publisher == nil {
		return payment, isDuplicate, nil
	}

	envelope, err := events.NewEnvelope("payment.completed", s.serviceName, traceID, map[string]any{
		"contract_id": payment.ContractID,
		"invoice_id":  payment.InvoiceID,
		"payment_id":  payment.ID,
		"amount":      payment.Amount,
		"paid_at":     payment.PaidAt,
	})
	if err != nil {
		s.logger.Error("failed to encode payment event", slog.String("error", err.Error()))
		return payment, false, nil
	}

	if err := s.publisher.Publish(ctx, "payment.completed", envelope); err != nil {
		s.logger.Error("failed to publish payment.completed", slog.String("error", err.Error()), slog.String("payment_id", payment.ID))
		return payment, false, nil
	}

	return payment, false, nil
}

func (s *PaymentService) GetPayment(ctx context.Context, id string) (domain.Payment, error) {
	return s.repo.GetPaymentByID(ctx, id)
}

func IsInvoiceNotFound(err error) bool {
	return errors.Is(err, repository.ErrInvoiceNotFound)
}

func IsPaymentNotFound(err error) bool {
	return errors.Is(err, repository.ErrPaymentNotFound)
}

func IsInvalidAmount(err error) bool {
	return errors.Is(err, repository.ErrInvalidAmount)
}
