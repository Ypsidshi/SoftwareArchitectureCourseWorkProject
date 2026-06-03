package httptransport

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"coursework/deal-service/internal/service"
	"coursework/platform-common/pkg/httpx"
	"github.com/go-chi/chi/v5"
)

// adminCancelBooking godoc
// @Summary Cancel booking (admin)
// @Tags admin-bookings
// @Produce json
// @Security BearerAuth
// @Param id path string true "Booking ID (UUID)"
// @Success 200 {object} domain.Booking
// @Failure 404 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Router /api/admin/bookings/{id} [delete]
func (h *Handler) adminCancelBooking(w http.ResponseWriter, r *http.Request) {
	bookingID := chi.URLParam(r, "id")
	booking, err := h.bookingService.AdminCancelBooking(r.Context(), httpx.TraceIDFromContext(r.Context()), bookingID)
	if err != nil {
		if service.IsBookingNotFound(err) {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"error": "booking not found"})
			return
		}
		h.logger.Error("admin cancel booking failed", slog.String("error", err.Error()))
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal error"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, booking)
}

// adminCheckoutBooking godoc
// @Summary Issue invoice for booking (admin)
// @Tags admin-bookings
// @Produce json
// @Security BearerAuth
// @Param id path string true "Booking ID (UUID)"
// @Success 200 {object} domain.Booking
// @Failure 404 {object} map[string]any
// @Failure 409 {object} map[string]any
// @Router /api/admin/bookings/{id}/checkout [post]
func (h *Handler) adminCheckoutBooking(w http.ResponseWriter, r *http.Request) {
	bookingID := chi.URLParam(r, "id")
	booking, err := h.bookingService.AdminCheckoutBooking(r.Context(), httpx.TraceIDFromContext(r.Context()), bookingID)
	if err != nil {
		writeBookingPaymentError(w, h, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, booking)
}

// adminPayBooking godoc
// @Summary Pay booking invoice (admin)
// @Tags admin-bookings
// @Produce json
// @Security BearerAuth
// @Param id path string true "Booking ID (UUID)"
// @Param Idempotency-Key header string false "Idempotency key"
// @Success 200 {object} service.PayBookingResult
// @Failure 404 {object} map[string]any
// @Router /api/admin/bookings/{id}/pay [post]
func (h *Handler) adminPayBooking(w http.ResponseWriter, r *http.Request) {
	bookingID := chi.URLParam(r, "id")
	result, err := h.bookingService.AdminPayBooking(r.Context(), httpx.TraceIDFromContext(r.Context()), bookingID, r.Header.Get("Idempotency-Key"))
	if err != nil {
		writeBookingPaymentError(w, h, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

func writeBookingPaymentError(w http.ResponseWriter, h *Handler, err error) {
	switch {
	case service.IsBookingNotFound(err):
		httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"error": "booking not found"})
	case errors.Is(err, service.ErrBookingNotPayable), errors.Is(err, service.ErrBookingAlreadyPaid):
		httpx.WriteJSON(w, http.StatusConflict, map[string]any{"error": err.Error()})
	case errors.Is(err, service.ErrInvoiceNotReady):
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
	default:
		h.logger.Error("booking payment failed", slog.String("error", err.Error()))
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
	}
}

// listContractsAdmin godoc
// @Summary List contracts (admin)
// @Tags admin-contracts
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param status query string false "Contract status"
// @Param payment_status query string false "Payment status"
// @Success 200 {object} service.ListContractsResult
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Router /api/admin/contracts [get]
func (h *Handler) listContractsAdmin(w http.ResponseWriter, r *http.Request) {
	page := intQueryOrDefault(r, "page", 1)
	pageSize := intQueryOrDefault(r, "page_size", 10)
	result, err := h.dealService.ListContracts(r.Context(), page, pageSize, service.AdminContractsFilter{
		PaymentStatus: strings.TrimSpace(r.URL.Query().Get("payment_status")),
		Status:        strings.TrimSpace(r.URL.Query().Get("status")),
	})
	if err != nil {
		h.logger.Error("list contracts failed", slog.String("error", err.Error()))
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal error"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

// payContractAdmin godoc
// @Summary Pay contract invoice (admin)
// @Tags admin-contracts
// @Produce json
// @Security BearerAuth
// @Param id path string true "Contract ID (UUID)"
// @Param Idempotency-Key header string false "Idempotency key"
// @Success 200 {object} domain.Contract
// @Failure 404 {object} map[string]any
// @Router /api/admin/contracts/{id}/pay [post]
func (h *Handler) payContractAdmin(w http.ResponseWriter, r *http.Request) {
	contractID := chi.URLParam(r, "id")
	contract, err := h.dealService.PayContract(r.Context(), httpx.TraceIDFromContext(r.Context()), contractID, r.Header.Get("Idempotency-Key"))
	if err != nil {
		switch {
		case service.IsNotFound(err):
			httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"error": "contract not found"})
		case errors.Is(err, service.ErrInvoiceNotReady):
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		default:
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, contract)
}

type sanatoriumBody struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	City            string   `json:"city"`
	Address         string   `json:"address"`
	DistanceToSeaKM float64  `json:"distance_to_sea_km"`
	Amenities       []string `json:"amenities"`
	ImageURLs       []string `json:"image_urls"`
	PricePerNight   float64  `json:"price_per_night"`
	TotalPlaces     int      `json:"total_places"`
	Latitude         *float64 `json:"latitude"`
	Longitude        *float64 `json:"longitude"`
	MedicalProfiles  []string `json:"medical_profiles"`
}

func decodeSanatoriumBody(r *http.Request) (service.SanatoriumInput, error) {
	var body sanatoriumBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return service.SanatoriumInput{}, err
	}
	return service.SanatoriumInput{
		Name: body.Name, Description: body.Description, City: body.City, Address: body.Address,
		DistanceToSeaKM: body.DistanceToSeaKM, Amenities: body.Amenities, ImageURLs: body.ImageURLs,
		PricePerNight: body.PricePerNight, TotalPlaces: body.TotalPlaces,
		Latitude: body.Latitude, Longitude: body.Longitude, MedicalProfiles: body.MedicalProfiles,
	}, nil
}

// listSanatoriumsAdmin godoc
// @Summary List sanatoriums (admin)
// @Tags admin-sanatoriums
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} service.ListSanatoriumsAdminResult
// @Router /api/admin/sanatoriums [get]
func (h *Handler) listSanatoriumsAdmin(w http.ResponseWriter, r *http.Request) {
	page := intQueryOrDefault(r, "page", 1)
	pageSize := intQueryOrDefault(r, "page_size", 20)
	result, err := h.bookingService.ListSanatoriumsAdmin(r.Context(), page, pageSize)
	if err != nil {
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal error"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

// createSanatoriumAdmin godoc
// @Summary Create sanatorium (admin)
// @Tags admin-sanatoriums
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body sanatoriumBody true "Sanatorium"
// @Success 201 {object} domain.Sanatorium
// @Failure 400 {object} map[string]any
// @Router /api/admin/sanatoriums [post]
func (h *Handler) createSanatoriumAdmin(w http.ResponseWriter, r *http.Request) {
	in, err := decodeSanatoriumBody(r)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	item, err := h.bookingService.CreateSanatorium(r.Context(), in)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, item)
}

// updateSanatoriumAdmin godoc
// @Summary Update sanatorium (admin)
// @Tags admin-sanatoriums
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Sanatorium ID (UUID)"
// @Param payload body sanatoriumBody true "Sanatorium"
// @Success 200 {object} domain.Sanatorium
// @Failure 404 {object} map[string]any
// @Router /api/admin/sanatoriums/{id} [put]
func (h *Handler) updateSanatoriumAdmin(w http.ResponseWriter, r *http.Request) {
	in, err := decodeSanatoriumBody(r)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	item, err := h.bookingService.UpdateSanatorium(r.Context(), in, chi.URLParam(r, "id"))
	if err != nil {
		if service.IsSanatoriumNotFound(err) {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"error": "sanatorium not found"})
			return
		}
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

// deleteSanatoriumAdmin godoc
// @Summary Delete sanatorium (admin)
// @Description Fails with 409 if active confirmed bookings exist.
// @Tags admin-sanatoriums
// @Produce json
// @Security BearerAuth
// @Param id path string true "Sanatorium ID (UUID)"
// @Success 200 {object} deletedResponse
// @Failure 404 {object} map[string]any
// @Failure 409 {object} map[string]any
// @Router /api/admin/sanatoriums/{id} [delete]
func (h *Handler) deleteSanatoriumAdmin(w http.ResponseWriter, r *http.Request) {
	err := h.bookingService.DeleteSanatorium(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		if service.IsSanatoriumNotFound(err) {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"error": "sanatorium not found"})
			return
		}
		if service.IsSanatoriumHasBookings(err) {
			httpx.WriteJSON(w, http.StatusConflict, map[string]any{"error": "sanatorium has active bookings"})
			return
		}
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal error"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true})
}
