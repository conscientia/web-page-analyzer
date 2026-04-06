package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// handleIndex serves the embedded frontend. Handler for GET /.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
}

// handleHealth responds with 200 OK. Handler for /health.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// handleAnalyze is the main endpoint. Handler for /analyze.
// Accepts POST with JSON body {"url": "https://..."} and returns
// either a PageAnalysis or an AnalysisError as JSON.
func (s *Server) handleAnalyze(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: expected {\"url\": \"...\"}")
		return
	}

	req.URL = strings.TrimSpace(req.URL)
	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "url is required")
		return
	}

	slog.Info("analyze request received", "url", req.URL)

	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.RequestTimeout)
	defer cancel()

	start := time.Now()
	result, analysisErr := s.analyzer.Analyze(ctx, req.URL)
	duration := time.Since(start)

	if analysisErr != nil {
		slog.Warn("analysis failed",
			"url", req.URL,
			"status", analysisErr.StatusCode,
			"error", analysisErr.Message,
			"duration", duration,
		)
		writeError(w, analysisErr.StatusCode, analysisErr.Message)
		return
	}

	slog.Info("analysis succeeded", "url", req.URL, "duration", duration)
	writeJSON(w, http.StatusOK, result)
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write JSON response", "error", err)
	}
}

// writeError writes a JSON error response using the AnalysisError shape.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, struct {
		StatusCode int    `json:"status_code"`
		Message    string `json:"message"`
	}{
		StatusCode: status,
		Message:    message,
	})
}
