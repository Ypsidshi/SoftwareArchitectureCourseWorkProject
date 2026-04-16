package httptransport

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"coursework/payment-service/internal/service"
	"coursework/platform-common/pkg/httpx"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handler struct {
	service        *service.PaymentService
	internalAPIKey string
	logger         *slog.Logger
}

func NewHandler(service *service.PaymentService, internalAPIKey string, logger *slog.Logger) *Handler {
	return &Handler{service: service, internalAPIKey: strings.TrimSpace(internalAPIKey), logger: logger}
}

func (h *Handler) Router(registry *prometheus.Registry) http.Handler {
	r := chi.NewRouter()
	r.Get("/health", h.ready)
	r.Get("/health/live", h.live)
	r.Get("/health/ready", h.ready)
	r.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	r.Post("/internal/invoices", h.createInvoice)

	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/payments", h.processPayment)
		api.Get("/payments/{id}", h.getPayment)
	})

	return r
}

type createInvoiceRequest struct {
	ContractID string  `json:"contract_id"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
}

func (h *Handler) createInvoice(w http.ResponseWriter, r *http.Request) {
	if h.internalAPIKey != "" && r.Header.Get("X-Internal-API-Key") != h.internalAPIKey {
		httpx.WriteJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid internal api key"})
		return
	}

	var req createInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}

	invoice, err := h.service.CreateInvoice(r.Context(), service.CreateInvoiceInput{
		ContractID: req.ContractID,
		Amount:     req.Amount,
		Currency:   req.Currency,
	})
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, map[string]any{
		"invoice_id": invoice.ID,
		"status":     invoice.Status,
		"invoice":    invoice,
	})
}

type processPaymentRequest struct {
	InvoiceID   string  `json:"invoice_id"`
	Amount      float64 `json:"amount"`
	ExternalRef string  `json:"external_ref"`
}

func (h *Handler) processPayment(w http.ResponseWriter, r *http.Request) {
	var req processPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}

	idempotencyKey := r.Header.Get("Idempotency-Key")
	payment, duplicate, err := h.service.ProcessPayment(r.Context(), httpx.TraceIDFromContext(r.Context()), service.ProcessPaymentInput{
		InvoiceID:      req.InvoiceID,
		Amount:         req.Amount,
		IdempotencyKey: idempotencyKey,
		ExternalRef:    req.ExternalRef,
	})
	if err != nil {
		switch {
		case service.IsInvoiceNotFound(err):
			httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"error": "invoice not found"})
		case service.IsInvalidAmount(err):
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid payment amount"})
		default:
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		}
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"payment":   payment,
		"duplicate": duplicate,
	})
}

func (h *Handler) getPayment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	payment, err := h.service.GetPayment(r.Context(), id)
	if err != nil {
		if service.IsPaymentNotFound(err) {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"error": "payment not found"})
			return
		}
		h.logger.Error("get payment failed", slog.String("error", err.Error()))
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal error"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, payment)
}

func (h *Handler) live(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "payment-service",
		"time":    time.Now().UTC(),
	})
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Ping(r.Context()); err != nil {
		httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "not_ready",
			"error":  "db unavailable",
		})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": "ready"})
}
