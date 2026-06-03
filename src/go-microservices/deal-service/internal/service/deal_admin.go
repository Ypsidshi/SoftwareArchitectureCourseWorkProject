package service

import (
	"context"
	"fmt"
	"strings"

	"coursework/deal-service/internal/domain"
	"coursework/deal-service/internal/integration/paymentclient"
	"coursework/deal-service/internal/repository"
)

type AdminContractsFilter struct {
	PaymentStatus string
	Status        string
}

type ListContractsResult struct {
	Items      []domain.Contract `json:"items"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

type contractListRepo interface {
	Ping(ctx context.Context) error
	CreateContract(ctx context.Context, c repository.NewContract) (domain.Contract, error)
	AttachInvoice(ctx context.Context, contractID, invoiceID string) error
	MarkInvoiceFailed(ctx context.Context, contractID, reason string) error
	UpdateStatus(ctx context.Context, contractID, status string) error
	UpdatePaymentStatus(ctx context.Context, contractID, paymentStatus string) error
	GetByID(ctx context.Context, id string) (domain.Contract, error)
	ListContracts(ctx context.Context, page, pageSize int, filter repository.AdminContractsFilter) ([]domain.Contract, int, error)
}

func (s *DealService) ListContracts(ctx context.Context, page, pageSize int, filter AdminContractsFilter) (ListContractsResult, error) {
	page, pageSize = normalizePagination(page, pageSize)
	repo, ok := s.repo.(contractListRepo)
	if !ok {
		return ListContractsResult{}, fmt.Errorf("contract list not supported")
	}
	items, total, err := repo.ListContracts(ctx, page, pageSize, repository.AdminContractsFilter{
		PaymentStatus: filter.PaymentStatus,
		Status:        filter.Status,
	})
	if err != nil {
		return ListContractsResult{}, err
	}
	totalPages := 0
	if pageSize > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	return ListContractsResult{
		Items: items, Total: total, Page: page, PageSize: pageSize, TotalPages: totalPages,
	}, nil
}

func (s *DealService) PayContract(ctx context.Context, traceID, contractID, idempotencyKey string) (domain.Contract, error) {
	contract, err := s.repo.GetByID(ctx, contractID)
	if err != nil {
		return domain.Contract{}, err
	}
	if contract.InvoiceID == nil || *contract.InvoiceID == "" {
		return domain.Contract{}, ErrInvoiceNotReady
	}
	if contract.PaymentStatus == "paid" {
		return contract, nil
	}

	gateway, ok := s.paymentClient.(bookingPaymentGateway)
	if !ok || gateway == nil {
		return domain.Contract{}, fmt.Errorf("payment client is not configured")
	}
	if strings.TrimSpace(idempotencyKey) == "" {
		idempotencyKey = fmt.Sprintf("contract-%s", contractID)
	}

	_, err = gateway.ProcessPayment(ctx, traceID, idempotencyKey, paymentclient.ProcessPaymentRequest{
		InvoiceID: *contract.InvoiceID,
		Amount:    contract.Amount,
	})
	if err != nil {
		return domain.Contract{}, err
	}
	_ = s.repo.UpdatePaymentStatus(ctx, contractID, "paid")
	return s.repo.GetByID(ctx, contractID)
}
