package analyzer

import (
	"context"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"
)

// checkAccessibility fires HTTP HEAD requests concurrently against all unique links.
// If HTTP HEAD is not allowed by server, falls back to GET
func (a *Analyzer) checkAccessibility(ctx context.Context, rawLinks []Link) Links {
	// split raw links into counts and unique links for checking
	stats, uniqueLinks := prepareLinks(rawLinks)

	if len(uniqueLinks) == 0 {
		return stats
	}

	var inaccessible atomic.Int64
	var unchecked atomic.Int64

	sem := make(chan struct{}, a.cfg.MaxWorkers)
	g, ctx := errgroup.WithContext(ctx)

	for _, link := range uniqueLinks {
		link := link

		g.Go(func() error {
			// acquire slot — blocks if MaxWorkers slots are taken
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				// link check time exhausted before we can start
				unchecked.Add(1)
				return nil
			}

			// per link timeout
			perLinkCtx, cancel := context.WithTimeout(ctx, a.cfg.PerLinkTimeout)
			defer cancel()

			if !a.isAccessible(perLinkCtx, link.URL) {
				inaccessible.Add(1)
			}

			return nil
		})
	}

	// wait for all workers to finish or context to be cancelled
	g.Wait()

	stats.Inaccessible = int(inaccessible.Load())
	stats.Unchecked = int(unchecked.Load())

	slog.Debug("link check complete",
		"total", stats.Total(),
		"internal", stats.Internal,
		"external", stats.External,
		"inaccessible", stats.Inaccessible,
		"unchecked", stats.Unchecked,
	)

	return stats
}

// prepareLinks splits rawLinks, which may contain duplicates into:
//   - a Links struct with Internal/External counts from all raw links
//   - a deduplicated slice of unique links for checking
func prepareLinks(rawLinks []Link) (Links, []Link) {
	var stats Links
	seen := make(map[string]bool)
	var unique []Link

	for _, link := range rawLinks {
		if link.IsInternal {
			stats.Internal++
		} else {
			stats.External++
		}

		if !seen[link.URL] {
			seen[link.URL] = true
			unique = append(unique, link)
		}
	}

	return stats, unique
}

// isAccessible sends a HEAD request to the given URL and returns true if
// the server responds with a status code below 400.
//
// Falls back to GET if the server returns 405 Method Not Allowed
func (a *Analyzer) isAccessible(ctx context.Context, rawURL string) bool {
	accessible, fallback := a.headCheck(ctx, rawURL)
	if fallback {
		// server does not support HEAD — retry with GET
		return a.getCheck(ctx, rawURL)
	}
	return accessible
}

// headCheck sends a HEAD request and returns whether we need to go for fallback GET
func (a *Analyzer) headCheck(ctx context.Context, rawURL string) (accessible bool, fallback bool) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, rawURL, nil)
	if err != nil {
		slog.Debug("head request build failed", "url", rawURL, "error", err)
		return false, false
	}

	req.Header.Set("User-Agent", "web-page-analyzer/1.0")

	start := time.Now()
	resp, err := a.client.Do(req)
	if err != nil {
		slog.Debug("head request failed", "url", rawURL, "error", err, "duration", time.Since(start))
		return false, false
	}
	defer resp.Body.Close()

	slog.Debug("head request complete", "url", rawURL, "status", resp.StatusCode, "duration", time.Since(start))

	if resp.StatusCode == http.StatusMethodNotAllowed {
		return false, true
	}

	return resp.StatusCode < http.StatusBadRequest, false
}

// getCheck sends a GET request and discards the body immediately.
// Used as a fallback when HEAD is not supported by the server.
func (a *Analyzer) getCheck(ctx context.Context, rawURL string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return false
	}

	req.Header.Set("User-Agent", "web-page-analyzer/1.0")

	start := time.Now()
	resp, err := a.client.Do(req)
	if err != nil {
		slog.Debug("get fallback failed", "url", rawURL, "error", err, "duration", time.Since(start))
		return false
	}
	defer resp.Body.Close()

	slog.Debug("get fallback complete", "url", rawURL, "status", resp.StatusCode, "duration", time.Since(start))

	return resp.StatusCode < http.StatusBadRequest
}
