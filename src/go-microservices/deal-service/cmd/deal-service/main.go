package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "coursework/deal-service/docs"
	"coursework/deal-service/internal/config"
	"coursework/deal-service/internal/integration/authclient"
	"coursework/deal-service/internal/integration/paymentclient"
	"coursework/deal-service/internal/repository"
	"coursework/deal-service/internal/service"
	eventstransport "coursework/deal-service/internal/transport/events"
	httptransport "coursework/deal-service/internal/transport/http"
	"coursework/platform-common/pkg/events"
	"coursework/platform-common/pkg/httpx"
	"coursework/platform-common/pkg/obs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// @title Sanatorium Booking API (deal-service)
// @version 1.1
// @description Public catalog and client bookings; auth proxy; admin API for bookings, contracts and sanatoriums. Payment is internal (deal → payment-service).
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT from POST /api/auth/login. Prefix: Bearer
func main() {
	cfg := config.Load()
	logger := obs.NewLogger(cfg.ServiceName, cfg.LogLevel)
	metrics := obs.NewHTTPMetrics(cfg.ServiceName)

	db, err := sql.Open("pgx", cfg.DBURL)
	if err != nil {
		logger.Error("failed to open db", slog.String("error", err.Error()))
		return
	}
	defer db.Close()
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	repo := repository.NewRepository(db)
	authClient := authclient.New(cfg.AuthServiceURL)
	paymentClient := paymentclient.New(cfg.PaymentServiceURL, cfg.InternalAPIKey)
	dealService := service.NewDealService(repo, paymentClient, logger)

	var publisher events.Publisher
	var natsConnClose func()
	natsConn, err := events.ConnectNATS(cfg.NATSURL)
	if err != nil {
		logger.Warn("nats is unavailable, async updates are disabled", slog.String("error", err.Error()))
	} else {
		publisher = events.NewNATSPublisher(natsConn)
		natsConnClose = natsConn.Close
	}
	if natsConnClose != nil {
		defer natsConnClose()
	}

	bookingService := service.NewBookingService(repo, paymentClient, publisher, cfg.ServiceName, logger)
	if natsConn != nil {
		paymentRouter := &service.PaymentEventsRouter{Deal: dealService, Booking: bookingService}
		if _, subErr := eventstransport.SubscribePaymentCompleted(natsConn, logger, paymentRouter); subErr != nil {
			logger.Warn("failed to subscribe payment.completed", slog.String("error", subErr.Error()))
		}
	}

	handler := httptransport.NewHandler(dealService, bookingService, authClient, cfg.JWTSecret, logger)

	router := handler.Router(metrics.Registry)
	chain := httpx.Chain(
		router,
		httpx.RecoveryMiddleware(logger),
		httpx.TraceMiddleware,
		httpx.LoggingMiddleware(logger),
		metrics.Middleware(cfg.ServiceName),
	)

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           chain,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("deal-service started", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server failed", slog.String("error", err.Error()))
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logger.Info("shutting down deal-service")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", slog.String("error", err.Error()))
		_ = srv.Close()
	}
}
