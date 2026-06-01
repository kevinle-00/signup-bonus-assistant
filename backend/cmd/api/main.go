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

	"github.com/kevinle-00/signup-bonus-assistant/backend/internal/api"
	"github.com/kevinle-00/signup-bonus-assistant/backend/internal/config"
	"github.com/kevinle-00/signup-bonus-assistant/backend/internal/db"
	"github.com/kevinle-00/signup-bonus-assistant/backend/internal/repositories"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	pool, err := db.OpenPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	cardOffers := repositories.NewPostgresCardOfferRepository(pool)
	recommendationRuns := repositories.NewPostgresRecommendationRunRepository(pool)
	handler := api.NewHandler(cardOffers, recommendationRuns)
	server := &http.Server{
		Addr:              cfg.APIAddr,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("api listening", "addr", cfg.APIAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen and serve", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown server", "error", err)
		os.Exit(1)
	}
}
