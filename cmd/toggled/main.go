package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/imabg/toggled/internal/api"
	"github.com/imabg/toggled/internal/config"
	"github.com/imabg/toggled/internal/store/postgres"
	"go.uber.org/zap"
)

// shutdownTimeout bounds graceful HTTP shutdown on SIGINT/SIGTERM.
const shutdownTimeout = 10 * time.Second

func main() {
	cfg, err := config.Load(config.DefaultPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "toggled: %v\n", err)
		os.Exit(1)
	}

	if _, err := config.NewLogger(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "toggled: %v\n", err)
		os.Exit(1)
	}
	defer zap.L().Sync() //nolint:errcheck

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.NewPool(ctx, cfg.PostgresURL())
	if err != nil {
		zap.L().Fatal("connect database", zap.Error(err))
	}
	defer pool.Close()

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:           api.NewRouter(pool),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			zap.L().Error("http shutdown", zap.Error(err))
		}
	}()

	zap.L().Info("server starting",
		zap.String("env", string(cfg.Env)),
		zap.Int("port", cfg.HTTP.Port),
	)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		zap.L().Fatal("http server", zap.Error(err))
	}
}
