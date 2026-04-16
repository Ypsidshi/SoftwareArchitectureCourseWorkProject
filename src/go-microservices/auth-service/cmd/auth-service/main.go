package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"coursework/auth-service/internal/config"
	"coursework/auth-service/internal/repository"
	"coursework/auth-service/internal/service"
	httptransport "coursework/auth-service/internal/transport/http"
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

	repo := repository.NewUserRepository(db)
	authService := service.NewAuthService(repo, cfg.JWTSecret, cfg.TokenTTL)
	handler := httptransport.NewHandler(authService, logger)

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
		logger.Info("auth-service started", slog.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server failed", slog.String("error", err.Error()))
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("shutting down auth-service")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", slog.String("error", err.Error()))
		_ = srv.Close()
	}
	fmt.Println("auth-service stopped")
}
