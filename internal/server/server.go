package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/conscientia/web-page-analyzer/internal/analyzer"
	"github.com/conscientia/web-page-analyzer/internal/config"
)

// Server holds the HTTP server and its dependencies.
type Server struct {
	http     *http.Server
	analyzer *analyzer.Analyzer
	cfg      *config.Config
}

// New creates a new Server.
// Routes /analyze, /health and / are configured.
func New(cfg *config.Config, a *analyzer.Analyzer) *Server {
	s := &Server{
		analyzer: a,
		cfg:      cfg,
	}

	mux := http.NewServeMux()

	// routes
	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("POST /analyze", s.handleAnalyze)
	mux.HandleFunc("GET /health", s.handleHealth)

	s.http = &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      s.withLogging(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: cfg.RequestTimeout + 5*time.Second, // must exceed analysis budget
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// Start starts the HTTP server and blocks until it is shut down.
func (s *Server) Start() error {
	slog.Info("server listening", "addr", s.http.Addr)

	if err := s.http.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Shutdown gracefully stops the server, waiting for in-flight requests
// to complete up to the given timeout.
func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("server shutting down")
	return s.http.Shutdown(ctx)
}

// withLogging is middleware that logs method, path, status, and duration
// for every request.
func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rw, r)

		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration", time.Since(start),
			"ip", r.RemoteAddr,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture the status code
// for logging middleware.
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}
