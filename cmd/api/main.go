package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jhatkaz/restaurant-console/internal/config"
	"github.com/jhatkaz/restaurant-console/internal/database"
	"github.com/jhatkaz/restaurant-console/internal/httpapi"
	"github.com/jhatkaz/restaurant-console/internal/repository"
	"github.com/jhatkaz/restaurant-console/internal/service"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil { logger.Error("database connection failed", "error", err); os.Exit(1) }
	defer db.Close()

	app := service.NewRestaurant(repository.New(db))
	server := &http.Server{Addr: cfg.HTTPAddr, Handler: httpapi.NewRouter(app, cfg.CORSOrigin, logger), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	go func() { logger.Info("api listening", "addr", cfg.HTTPAddr); if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) { logger.Error("server stopped", "error", err); stop() } }()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second); defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil { logger.Error("graceful shutdown failed", "error", err) }
}
