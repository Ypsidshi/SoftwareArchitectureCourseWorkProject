package paymentclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type CreateInvoiceRequest struct {
	ContractID string  `json:"contract_id,omitempty"`
	BookingID  string  `json:"booking_id,omitempty"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
}

type ProcessPaymentRequest struct {
	InvoiceID   string  `json:"invoice_id"`
	Amount      float64 `json:"amount"`
	ExternalRef string  `json:"external_ref,omitempty"`
}

type ProcessPaymentResponse struct {
	Payment   map[string]any `json:"payment"`
	Duplicate bool           `json:"duplicate"`
}

type CreateInvoiceResponse struct {
	InvoiceID string `json:"invoice_id"`
	Status    string `json:"status"`
}

type Client struct {
	baseURL        string
	internalAPIKey string
	client         *http.Client
}

func New(baseURL, internalAPIKey string) *Client {
	return &Client{
		baseURL:        strings.TrimRight(baseURL, "/"),
		internalAPIKey: strings.TrimSpace(internalAPIKey),
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (c *Client) CreateInvoice(ctx context.Context, traceID string, req CreateInvoiceRequest) (CreateInvoiceResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return CreateInvoiceResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/invoices", bytes.NewBuffer(body))
	if err != nil {
		return CreateInvoiceResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if traceID != "" {
		httpReq.Header.Set("X-Trace-Id", traceID)
	}
	if c.internalAPIKey != "" {
		httpReq.Header.Set("X-Internal-API-Key", c.internalAPIKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return CreateInvoiceResponse{}, err
	}
	defer resp.Body.Close()

	payload, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		return CreateInvoiceResponse{}, fmt.Errorf("payment service returned %d: %s", resp.StatusCode, string(payload))
	}

	var out CreateInvoiceResponse
	if err := json.Unmarshal(payload, &out); err != nil {
		return CreateInvoiceResponse{}, fmt.Errorf("decode create invoice response: %w", err)
	}
	return out, nil
}

func (c *Client) ProcessPayment(ctx context.Context, traceID, idempotencyKey string, req ProcessPaymentRequest) (ProcessPaymentResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return ProcessPaymentResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/payments", bytes.NewBuffer(body))
	if err != nil {
		return ProcessPaymentResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if traceID != "" {
		httpReq.Header.Set("X-Trace-Id", traceID)
	}
	if idempotencyKey != "" {
		httpReq.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return ProcessPaymentResponse{}, err
	}
	defer resp.Body.Close()

	payload, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= http.StatusBadRequest {
		return ProcessPaymentResponse{}, fmt.Errorf("payment service returned %d: %s", resp.StatusCode, string(payload))
	}

	var out struct {
		Payment   map[string]any `json:"payment"`
		Duplicate bool           `json:"duplicate"`
	}
	if err := json.Unmarshal(payload, &out); err != nil {
		return ProcessPaymentResponse{}, fmt.Errorf("decode process payment response: %w", err)
	}
	return ProcessPaymentResponse{Payment: out.Payment, Duplicate: out.Duplicate}, nil
}
