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

	"coursework/payment-service/internal/config"
	"coursework/payment-service/internal/repository"
	"coursework/payment-service/internal/service"
	httptransport "coursework/payment-service/internal/transport/http"
	"coursework/platform-common/pkg/events"
	"coursework/platform-common/pkg/httpx"
	"coursework/platform-common/pkg/obs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

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

	var publisher events.Publisher
	natsConn, err := events.ConnectNATS(cfg.NATSURL)
	if err != nil {
		logger.Warn("nats is unavailable, events will not be published", slog.String("error", err.Error()))
	} else {
		publisher = events.NewNATSPublisher(natsConn)
		defer natsConn.Close()
	}

	repo := repository.NewRepository(db)
	paymentService := service.NewPaymentService(repo, publisher, cfg.ServiceName, logger)
	handler := httptransport.NewHandler(paymentService, cfg.InternalAPIKey, logger)

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
		logger.Info("payment-service started", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server failed", slog.String("error", err.Error()))
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logger.Info("shutting down payment-service")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", slog.String("error", err.Error()))
		_ = srv.Close()
	}
}
