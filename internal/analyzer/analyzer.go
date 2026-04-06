package analyzer

import (
	"context"
	"log/slog"
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
	slog.Info("analysis started", "url", rawURL)

	// step 1: validate and parse the URL
	baseURL, analysisErr := parseURL(rawURL)
	if analysisErr != nil {
		slog.Warn("invalid URL", "url", rawURL, "error", analysisErr.Message)
		return nil, analysisErr
	}

	// step 2: fetch the page
	fetchCtx, cancel := context.WithTimeout(ctx, a.cfg.FetchTimeout)
	defer cancel()

	body, finalURL, analysisErr := a.fetchPage(fetchCtx, baseURL)
	if analysisErr != nil {
		slog.Warn("fetch failed", "url", rawURL, "error", analysisErr.Message)
		return nil, analysisErr
	}

	// step 3: parse HTML into node tree
	doc, err := parseHTML(body)
	if err != nil {
		slog.Error("html parse failed", "url", rawURL, "error", err)
		return nil, &AnalysisError{
			StatusCode: http.StatusInternalServerError,
			Message:    "failed to parse HTML",
		}
	}

	// step 4: extract base URL
	effectiveBase := extractBaseURL(doc, finalURL)

	// step 5: extract information
	htmlVersion := detectHTMLVersion(body)
	title := extractTitle(doc)
	headings := countHeadings(doc)
	hasLoginForm := detectLoginForm(doc)
	rawLinks := extractLinks(doc, effectiveBase)

	// step 6: link accessibility concurrently
	linkCtx, linkCancel := context.WithTimeout(ctx, a.cfg.LinkCheckTimeout)
	defer linkCancel()

	links := a.checkAccessibility(linkCtx, rawLinks)

	slog.Info("analysis complete",
		"url", rawURL,
		"htmlVersion", htmlVersion,
		"title", title,
		"links", links.Total(),
		"inaccessible", links.Inaccessible,
		"unchecked", links.Unchecked,
		"hasLoginForm", hasLoginForm,
	)

	return &PageAnalysis{
		URL:          finalURL.String(),
		HTMLVersion:  htmlVersion,
		Title:        title,
		Headings:     headings,
		Links:        links,
		HasLoginForm: hasLoginForm,
	}, nil
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
