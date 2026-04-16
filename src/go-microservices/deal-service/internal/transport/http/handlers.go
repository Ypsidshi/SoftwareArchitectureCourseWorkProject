package httptransport

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"coursework/deal-service/internal/integration/authclient"
	"coursework/deal-service/internal/service"
	"coursework/platform-common/pkg/httpx"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Handler struct {
	dealService    *service.DealService
	bookingService *service.BookingService
	authClient     *authclient.Client
	jwtSecret      string
	logger         *slog.Logger
}

func NewHandler(dealService *service.DealService, bookingService *service.BookingService, authClient *authclient.Client, jwtSecret string, logger *slog.Logger) *Handler {
	return &Handler{
		dealService:    dealService,
		bookingService: bookingService,
		authClient:     authClient,
		jwtSecret:      jwtSecret,
		logger:         logger,
	}
}

func (h *Handler) Router(registry *prometheus.Registry) http.Handler {
	r := chi.NewRouter()
	r.Get("/health", h.ready)
	r.Get("/health/live", h.live)
	r.Get("/health/ready", h.ready)
	r.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DeepLinking(true),
	))

	r.Route("/api", func(api chi.Router) {
		api.Post("/auth/login", h.loginViaAuthService)
		api.Get("/sanatoriums", h.listSanatoriums)
		api.Get("/sanatoriums/{id}", h.getSanatoriumByID)

		api.Group(func(authorized chi.Router) {
			authorized.Use(ClientAuthMiddleware(h.jwtSecret, h.logger))
			authorized.Post("/bookings", h.createBooking)
			authorized.Get("/bookings", h.listBookings)
			authorized.Get("/bookings/{id}", h.getBookingByID)
			authorized.Put("/bookings/{id}", h.updateBooking)
			authorized.Delete("/bookings/{id}", h.cancelBooking)
		})
	})

	r.Route("/api/v1", func(api chi.Router) {
		api.Group(func(authorized chi.Router) {
			authorized.Use(AuthMiddleware(h.jwtSecret, h.logger, "admin", "manager", "accountant"))
			authorized.Post("/contracts", h.createContract)
			authorized.Get("/contracts/{id}", h.getContract)
			authorized.Patch("/contracts/{id}/status", h.updateStatus)
		})
	})
	return r
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginViaAuthService godoc
// @Summary Login and get JWT token
// @Description Proxies login request to auth-service and returns access token.
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body loginRequest true "Login payload"
// @Success 200 {object} map[string]any
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/auth/login [post]
func (h *Handler) loginViaAuthService(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "email and password are required"})
		return
	}

	resp, err := h.authClient.Login(r.Context(), httpx.TraceIDFromContext(r.Context()), authclient.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		errText := err.Error()
		if strings.Contains(errText, " 401: ") || strings.Contains(errText, " 400: ") {
			httpx.WriteJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid credentials"})
			return
		}
		h.logger.Error("auth login proxy failed", slog.String("error", errText))
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to authenticate"})
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"access_token": resp.AccessToken,
		"user":         resp.User,
	})
}

type createContractRequest struct {
	ResidentID string  `json:"resident_id"`
	RoomID     string  `json:"room_id"`
	ManagerID  string  `json:"manager_id"`
	StartDate  string  `json:"start_date"`
	EndDate    string  `json:"end_date"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
}

func (h *Handler) createContract(w http.ResponseWriter, r *http.Request) {
	var req createContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "start_date format must be YYYY-MM-DD"})
		return
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "end_date format must be YYYY-MM-DD"})
		return
	}

	contract, err := h.dealService.CreateContract(r.Context(), httpx.TraceIDFromContext(r.Context()), service.CreateContractInput{
		ResidentID: req.ResidentID,
		RoomID:     req.RoomID,
		ManagerID:  req.ManagerID,
		StartDate:  startDate,
		EndDate:    endDate,
		Amount:     req.Amount,
		Currency:   req.Currency,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvoiceFailure) {
			httpx.WriteJSON(w, http.StatusBadGateway, map[string]any{
				"error":    "contract created but invoice creation failed",
				"contract": contract,
			})
			return
		}
		h.logger.Error("create contract failed", slog.String("error", err.Error()), slog.String("trace_id", httpx.TraceIDFromContext(r.Context())))
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, contract)
}

func (h *Handler) getContract(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	item, err := h.dealService.GetContract(r.Context(), id)
	if err != nil {
		if service.IsNotFound(err) {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"error": "contract not found"})
			return
		}
		h.logger.Error("get contract failed", slog.String("error", err.Error()))
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal error"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}

	if err := h.dealService.UpdateStatus(r.Context(), id, req.Status); err != nil {
		if service.IsNotFound(err) {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"error": "contract not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidStatus) {
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid status"})
			return
		}
		h.logger.Error("update status failed", slog.String("error", err.Error()))
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal error"})
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": "updated"})
}

func (h *Handler) live(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "deal-service",
		"time":    time.Now().UTC(),
	})
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	if err := h.dealService.Ping(r.Context()); err != nil {
		httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "not_ready",
			"error":  "db unavailable",
		})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"status": "ready"})
}

// listSanatoriums godoc
// @Summary List sanatoriums
// @Description Returns sanatorium catalog with pagination, filtering and sorting.
// @Tags sanatoriums
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param city query string false "City filter"
// @Param profiles query string false "Medical profiles (CSV), example: cardiology,pulmonology"
// @Param max_distance_to_sea query number false "Maximum distance to sea in km"
// @Param price_min query number false "Minimum price per night"
// @Param price_max query number false "Maximum price per night"
// @Param check_in query string false "Availability start date (YYYY-MM-DD)"
// @Param check_out query string false "Availability end date (YYYY-MM-DD)"
// @Param sort query string false "Sort mode: price_asc,price_desc,distance_asc,distance_desc"
// @Success 200 {object} service.ListSanatoriumsResult
// @Failure 400 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/sanatoriums [get]
func (h *Handler) listSanatoriums(w http.ResponseWriter, r *http.Request) {
	page := intQueryOrDefault(r, "page", 1)
	pageSize := intQueryOrDefault(r, "page_size", 10)
	city := strings.TrimSpace(r.URL.Query().Get("city"))
	sort := strings.TrimSpace(r.URL.Query().Get("sort"))
	profiles := splitCSV(r.URL.Query().Get("profiles"))

	maxDistance, err := optionalFloatQuery(r, "max_distance_to_sea")
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	priceMin, err := optionalFloatQuery(r, "price_min")
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	priceMax, err := optionalFloatQuery(r, "price_max")
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	checkIn, err := optionalDateQuery(r, "check_in")
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	checkOut, err := optionalDateQuery(r, "check_out")
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	result, err := h.bookingService.ListSanatoriums(r.Context(), service.ListSanatoriumsInput{
		Page:               page,
		PageSize:           pageSize,
		City:               city,
		ProfileNames:       profiles,
		MaxDistanceToSeaKM: maxDistance,
		PriceMin:           priceMin,
		PriceMax:           priceMax,
		CheckIn:            checkIn,
		CheckOut:           checkOut,
		Sort:               sort,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidDateRange) {
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid check_in/check_out date range"})
			return
		}
		h.logger.Error("list sanatoriums failed", slog.String("error", err.Error()))
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal error"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

// getSanatoriumByID godoc
// @Summary Get sanatorium details
// @Description Returns detailed information for one sanatorium.
// @Tags sanatoriums
// @Produce json
// @Param id path string true "Sanatorium ID (UUID)"
// @Param check_in query string false "Availability start date (YYYY-MM-DD)"
// @Param check_out query string false "Availability end date (YYYY-MM-DD)"
// @Success 200 {object} service.SanatoriumDetailsResult
// @Failure 400 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/sanatoriums/{id} [get]
func (h *Handler) getSanatoriumByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	checkIn, err := optionalDateQuery(r, "check_in")
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	checkOut, err := optionalDateQuery(r, "check_out")
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if (checkIn != nil && checkOut == nil) || (checkIn == nil && checkOut != nil) {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "both check_in and check_out must be provided together"})
		return
	}

	result, err := h.bookingService.GetSanatoriumDetails(r.Context(), id, checkIn, checkOut)
	if err != nil {
		switch {
		case service.IsSanatoriumNotFound(err):
			httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"error": "sanatorium not found"})
		case errors.Is(err, service.ErrInvalidDateRange):
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid check_in/check_out date range"})
		default:
			h.logger.Error("get sanatorium failed", slog.String("error", err.Error()))
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal error"})
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

type createBookingRequest struct {
	SanatoriumID string `json:"sanatorium_id"`
	CheckIn      string `json:"check_in"`
	CheckOut     string `json:"check_out"`
	Guests       int    `json:"guests"`
}

// createBooking godoc
// @Summary Create booking
// @Description Creates booking for authorized client (role=client).
// @Tags bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body createBookingRequest true "Booking payload"
// @Success 201 {object} domain.Booking
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 409 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/bookings [post]
func (h *Handler) createBooking(w http.ResponseWriter, r *http.Request) {
	clientID := ClientIDFromContext(r.Context())

	var req createBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "check_in format must be YYYY-MM-DD"})
		return
	}
	checkOut, err := time.Parse("2006-01-02", req.CheckOut)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "check_out format must be YYYY-MM-DD"})
		return
	}

	booking, err := h.bookingService.CreateBooking(r.Context(), httpx.TraceIDFromContext(r.Context()), service.CreateBookingInput{
		ClientID:     clientID,
		SanatoriumID: req.SanatoriumID,
		CheckIn:      checkIn,
		CheckOut:     checkOut,
		Guests:       req.Guests,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidDateRange), errors.Is(err, service.ErrInvalidGuests):
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		case service.IsSanatoriumNotFound(err):
			httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"error": "sanatorium not found"})
		case service.IsSanatoriumNotAvailable(err):
			httpx.WriteJSON(w, http.StatusConflict, map[string]any{"error": "selected sanatorium is not available on these dates"})
		case service.IsGuestsExceedCapacity(err):
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "guests exceed sanatorium capacity"})
		default:
			h.logger.Error("create booking failed", slog.String("error", err.Error()))
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal error"})
		}
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, booking)
}

// listBookings godoc
// @Summary List current client bookings
// @Description Returns booking list for authorized client.
// @Tags bookings
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} service.ListBookingsResult
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/bookings [get]
func (h *Handler) listBookings(w http.ResponseWriter, r *http.Request) {
	clientID := ClientIDFromContext(r.Context())
	page := intQueryOrDefault(r, "page", 1)
	pageSize := intQueryOrDefault(r, "page_size", 10)

	result, err := h.bookingService.ListBookings(r.Context(), clientID, page, pageSize)
	if err != nil {
		h.logger.Error("list bookings failed", slog.String("error", err.Error()))
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal error"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

// getBookingByID godoc
// @Summary Get booking details
// @Description Returns one booking of authorized client.
// @Tags bookings
// @Produce json
// @Security BearerAuth
// @Param id path string true "Booking ID (UUID)"
// @Success 200 {object} domain.Booking
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/bookings/{id} [get]
func (h *Handler) getBookingByID(w http.ResponseWriter, r *http.Request) {
	clientID := ClientIDFromContext(r.Context())
	id := chi.URLParam(r, "id")

	booking, err := h.bookingService.GetBooking(r.Context(), id, clientID)
	if err != nil {
		if service.IsBookingNotFound(err) {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"error": "booking not found"})
			return
		}
		h.logger.Error("get booking failed", slog.String("error", err.Error()))
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal error"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, booking)
}

type updateBookingRequest struct {
	CheckIn  string `json:"check_in"`
	CheckOut string `json:"check_out"`
	Guests   int    `json:"guests"`
}

// updateBooking godoc
// @Summary Update booking
// @Description Updates booking dates and guests for authorized client.
// @Tags bookings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Booking ID (UUID)"
// @Param payload body updateBookingRequest true "Update booking payload"
// @Success 200 {object} domain.Booking
// @Failure 400 {object} map[string]any
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 409 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/bookings/{id} [put]
func (h *Handler) updateBooking(w http.ResponseWriter, r *http.Request) {
	clientID := ClientIDFromContext(r.Context())
	bookingID := chi.URLParam(r, "id")

	var req updateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	checkIn, err := time.Parse("2006-01-02", req.CheckIn)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "check_in format must be YYYY-MM-DD"})
		return
	}
	checkOut, err := time.Parse("2006-01-02", req.CheckOut)
	if err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "check_out format must be YYYY-MM-DD"})
		return
	}

	booking, err := h.bookingService.UpdateBooking(r.Context(), httpx.TraceIDFromContext(r.Context()), service.UpdateBookingInput{
		BookingID: bookingID,
		ClientID:  clientID,
		CheckIn:   checkIn,
		CheckOut:  checkOut,
		Guests:    req.Guests,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidDateRange), errors.Is(err, service.ErrInvalidGuests):
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
		case service.IsBookingNotFound(err):
			httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"error": "booking not found"})
		case service.IsSanatoriumNotAvailable(err):
			httpx.WriteJSON(w, http.StatusConflict, map[string]any{"error": "selected sanatorium is not available on these dates"})
		case service.IsGuestsExceedCapacity(err):
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "guests exceed sanatorium capacity"})
		default:
			h.logger.Error("update booking failed", slog.String("error", err.Error()))
			httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal error"})
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, booking)
}

// cancelBooking godoc
// @Summary Cancel booking
// @Description Cancels booking of authorized client.
// @Tags bookings
// @Produce json
// @Security BearerAuth
// @Param id path string true "Booking ID (UUID)"
// @Success 200 {object} domain.Booking
// @Failure 401 {object} map[string]any
// @Failure 403 {object} map[string]any
// @Failure 404 {object} map[string]any
// @Failure 500 {object} map[string]any
// @Router /api/bookings/{id} [delete]
func (h *Handler) cancelBooking(w http.ResponseWriter, r *http.Request) {
	clientID := ClientIDFromContext(r.Context())
	bookingID := chi.URLParam(r, "id")

	booking, err := h.bookingService.CancelBooking(r.Context(), httpx.TraceIDFromContext(r.Context()), bookingID, clientID)
	if err != nil {
		if service.IsBookingNotFound(err) {
			httpx.WriteJSON(w, http.StatusNotFound, map[string]any{"error": "booking not found"})
			return
		}
		h.logger.Error("cancel booking failed", slog.String("error", err.Error()))
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal error"})
		return
	}
	httpx.WriteJSON(w, http.StatusOK, booking)
}

func intQueryOrDefault(r *http.Request, key string, fallback int) int {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}

func optionalFloatQuery(r *http.Request, key string) (*float64, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, fmt.Errorf("%s must be a number", key)
	}
	return &value, nil
}

func optionalDateQuery(r *http.Request, key string) (*time.Time, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, fmt.Errorf("%s format must be YYYY-MM-DD", key)
	}
	return &t, nil
}

func splitCSV(value string) []string {
	parts := strings.Split(strings.TrimSpace(value), ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
