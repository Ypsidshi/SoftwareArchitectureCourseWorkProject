package httptransport

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"coursework/auth-service/internal/service"
	"coursework/platform-common/pkg/httpx"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handler struct {
	auth   *service.AuthService
	logger *slog.Logger
}

func NewHandler(auth *service.AuthService, logger *slog.Logger) *Handler {
	return &Handler{auth: auth, logger: logger}
}

func (h *Handler) Router(registry *prometheus.Registry) http.Handler {
	r := chi.NewRouter()

	r.Get("/health", h.ready)
	r.Get("/health/live", h.live)
	r.Get("/health/ready", h.ready)
	r.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	r.Route("/api/v1", func(api chi.Router) {
		api.Post("/users/register", h.register)
		api.Post("/auth/login", h.login)
	})

	return r
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" || strings.TrimSpace(req.FullName) == "" || strings.TrimSpace(req.Role) == "" {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "email, password, full_name and role are required"})
		return
	}

	user, err := h.auth.Register(r.Context(), req.Email, req.Password, req.FullName, req.Role)
	if err != nil {
		if service.IsEmailExists(err) {
			httpx.WriteJSON(w, http.StatusConflict, map[string]any{"error": "email already exists"})
			return
		}
		if errors.Is(err, service.ErrInvalidEmail) {
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "email must be valid"})
			return
		}
		if errors.Is(err, service.ErrWeakPassword) {
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "password must be at least 8 characters long"})
			return
		}
		if errors.Is(err, service.ErrInvalidFullName) {
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "full_name is required"})
			return
		}
		if errors.Is(err, service.ErrInvalidRole) {
			httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "public registration is available only for role=client"})
			return
		}
		h.logger.Error("register failed", slog.String("error", err.Error()), slog.String("trace_id", httpx.TraceIDFromContext(r.Context())))
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal error"})
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, user)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}

	token, user, err := h.auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			httpx.WriteJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid credentials"})
			return
		}
		h.logger.Error("login failed", slog.String("error", err.Error()), slog.String("trace_id", httpx.TraceIDFromContext(r.Context())))
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]any{"error": "internal error"})
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"access_token": token,
		"user":         user,
	})
}

func (h *Handler) live(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "auth-service",
		"time":    time.Now().UTC(),
	})
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	if err := h.auth.Ping(r.Context()); err != nil {
		httpx.WriteJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "not_ready",
			"error":  "db unavailable",
		})
		return
	}

	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status": "ready",
	})
}
