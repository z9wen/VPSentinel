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

	"github.com/gin-gonic/gin"

	"github.com/z9wen/VPSentinel/server/internal/api"
	"github.com/z9wen/VPSentinel/server/internal/auth"
	"github.com/z9wen/VPSentinel/server/internal/config"
	"github.com/z9wen/VPSentinel/server/internal/database"
	"github.com/z9wen/VPSentinel/server/internal/version"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	if err := run(logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	databaseCtx, cancelDatabaseStartup := context.WithTimeout(ctx, cfg.DatabaseStartTimeout)
	pool, err := database.Connect(databaseCtx, cfg.DatabaseURL, logger)
	if err != nil {
		cancelDatabaseStartup()
		return err
	}
	defer pool.Close()
	if cfg.AutoMigrate {
		if err := database.Migrate(databaseCtx, pool, logger); err != nil {
			cancelDatabaseStartup()
			return err
		}
	} else {
		logger.Info("automatic database migrations disabled")
	}
	cancelDatabaseStartup()

	gin.SetMode(cfg.GinMode)
	store := database.NewPostgresStore(pool)
	router := api.NewRouter(api.Dependencies{
		Authenticator: auth.NewStatic(cfg.AgentToken, cfg.NodeID),
		Store:         store,
		Logger:        logger,
		Version:       version.Version,
		MaxBodyBytes:  cfg.MaxBodyBytes,
	})

	server := &http.Server{
		Addr:              cfg.ListenAddress,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("server listening",
			"address", cfg.ListenAddress,
			"version", version.Version,
			"storage", "postgresql",
		)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}
	logger.Info("server shut down gracefully")
	return nil
}
