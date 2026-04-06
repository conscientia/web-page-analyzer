package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/conscientia/web-page-analyzer/internal/analyzer"
	"github.com/conscientia/web-page-analyzer/internal/config"
	"github.com/conscientia/web-page-analyzer/internal/logger"
	"github.com/conscientia/web-page-analyzer/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	logger.Setup(cfg.LogLevel, cfg.LogFormat)

	slog.Info("configuration loaded", "config", cfg.String())

	analyzer := analyzer.New(cfg)
	srv := server.New(cfg, analyzer)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// waits for signal then drains in-flight requests
	go func() {
		<-ctx.Done()
		slog.Info("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("shutdown error", "error", err)
		}
	}()

	// blocks until server is shut down
	if err := srv.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}

	slog.Info("server stopped")
}
