package analyzer

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
)

// fetchPage fetches the page at the given URL and returns the raw body bytes
// and the final URL after any redirects.
func (a *Analyzer) fetchPage(ctx context.Context, target *url.URL) ([]byte, *url.URL, *AnalysisError) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, nil, &AnalysisError{
			StatusCode: http.StatusBadRequest,
			Message:    fmt.Sprintf("could not build request: %v", err),
		}
	}

	req.Header.Set("User-Agent", "web-page-analyzer/1.0")

	slog.Info("fetching page", "url", target.String())

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, nil, &AnalysisError{
			StatusCode: http.StatusBadGateway,
			Message:    fetchErrorMessage(err),
		}
	}
	defer resp.Body.Close()

	// surface non-2xx responses as errors — the page exists but is not usable
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, nil, &AnalysisError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("server returned %d %s", resp.StatusCode, http.StatusText(resp.StatusCode)),
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, &AnalysisError{
			StatusCode: http.StatusInternalServerError,
			Message:    fmt.Sprintf("could not read response body: %v", err),
		}
	}

	// use the final URL after redirects as the base for link classification
	finalURL := resp.Request.URL

	slog.Debug("page fetched",
		"url", finalURL.String(),
		"status", resp.StatusCode,
		"bytes", len(body),
	)

	return body, finalURL, nil
}

// fetchErrorMessage converts a client.Do error into a user-friendly message.
func fetchErrorMessage(err error) string {
	// FETCH_TIMEOUT expired
	if errors.Is(err, context.DeadlineExceeded) {
		return "request timed out: the server took too long to respond"
	}

	// browser disconnected or parent context was cancelled
	if errors.Is(err, context.Canceled) {
		return "request was cancelled"
	}

	// DNS failure — host not found
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return fmt.Sprintf("could not resolve host %q: DNS lookup failed", dnsErr.Name)
	}

	// connection refused or network unreachable — server is down
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return fmt.Sprintf("could not connect to server: %v", opErr.Err)
	}

	// TLS certificate issued by unknown authority
	var certErr x509.UnknownAuthorityError
	if errors.As(err, &certErr) {
		return "TLS certificate verification failed: untrusted certificate authority"
	}

	// TLS certificate is for a different hostname
	var hostErr x509.HostnameError
	if errors.As(err, &hostErr) {
		return fmt.Sprintf("TLS certificate hostname mismatch: certificate is for %q", hostErr.Certificate.Subject.CommonName)
	}

	// fallback — surface the raw error
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return fmt.Sprintf("could not reach URL: %v", urlErr.Err)
	}

	return fmt.Sprintf("could not reach URL: %v", err)
}
