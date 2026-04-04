package main

import (
	"github.com/conscientia/web-page-analyzer/internal/config"
	"github.com/conscientia/web-page-analyzer/internal/logger"
	"log"
	"log/slog"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	logger.Setup(cfg.LogLevel, cfg.LogFormat)

	slog.Info("starting web-page-analyzer")
}
