package analyzer

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/conscientia/web-page-analyzer/internal/config"
)

// Analyzer orchestrates the full page analysis.
// It holds shared dependencies — the HTTP client and config —
// so they are created once at startup and reused across requests.
type Analyzer struct {
	client *http.Client
	cfg    *config.Config
}

// New creates a new Analyzer.
func New(cfg *config.Config) *Analyzer {
	return &Analyzer{
		client: &http.Client{
			Timeout: cfg.FetchTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     30 * time.Second,
			},
		},
		cfg: cfg,
	}
}

// Analyze fetches the page at the given rawURL and returns a PageAnalysis.
// Returns AnalysisError if the URL is invalid or the page is unreachable.
func (a *Analyzer) Analyze(ctx context.Context, rawURL string) (*PageAnalysis, *AnalysisError) {
	// step 1: validate and parse the URL
	_, analysisErr := parseURL(rawURL)
	if analysisErr != nil {
		return nil, analysisErr
	}

	// step 2: fetch the page
	// step 3: parse HTML
	// step 4: extract base URL
	// step 5: title, headings, links
	// step 6: link accessibility

	return nil, nil // placeholder until steps are implemented
}

// parseURL validates and parses the raw URL string.
// Only http and https schemes are accepted.
func parseURL(rawURL string) (*url.URL, *AnalysisError) {
	if rawURL == "" {
		return nil, &AnalysisError{
			StatusCode: http.StatusBadRequest,
			Message:    "URL is required",
		}
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, &AnalysisError{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid URL format",
		}
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, &AnalysisError{
			StatusCode: http.StatusBadRequest,
			Message:    "unsupported scheme: only http and https are allowed",
		}
	}

	if parsed.Host == "" {
		return nil, &AnalysisError{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid URL: missing host",
		}
	}

	return parsed, nil
}
