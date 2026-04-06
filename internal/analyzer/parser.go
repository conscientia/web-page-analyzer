package analyzer

import (
	"bytes"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// parseHTML parses raw HTML bytes into a node tree.
func parseHTML(body []byte) (*html.Node, error) {
	return html.Parse(bytes.NewReader(body))
}

// detectHTMLVersion inspects the raw DOCTYPE declaration
// and returns a human-readable HTML version string.
func detectHTMLVersion(body []byte) string {
	// read only the first 512 bytes — DOCTYPE is always at the top
	// Node does not have this information, so using this methodology
	preview := strings.ToLower(string(body[:min(512, len(body))]))

	switch {
	case strings.Contains(preview, "<!doctype html>"):
		return "HTML5"
	case strings.Contains(preview, "xhtml 1.1"):
		return "XHTML 1.1"
	case strings.Contains(preview, "xhtml 1.0 strict"):
		return "XHTML 1.0 Strict"
	case strings.Contains(preview, "xhtml 1.0 transitional"):
		return "XHTML 1.0 Transitional"
	case strings.Contains(preview, "xhtml 1.0 frameset"):
		return "XHTML 1.0 Frameset"
	case strings.Contains(preview, "html 4.01 strict"):
		return "HTML 4.01 Strict"
	case strings.Contains(preview, "html 4.01 transitional"):
		return "HTML 4.01 Transitional"
	case strings.Contains(preview, "html 4.01 frameset"):
		return "HTML 4.01 Frameset"
	case strings.Contains(preview, "html 4.0"):
		return "HTML 4.0"
	case strings.Contains(preview, "html 3.2"):
		return "HTML 3.2"
	case strings.Contains(preview, "html 2.0"):
		return "HTML 2.0"
	case strings.Contains(preview, "<!doctype"):
		return "Unknown DOCTYPE"
	default:
		return "Unknown"
	}
}

// extractBaseURL checks for a href tag in the document head.
// If present, relative links are resolved against it rather than the fetched URL.
func extractBaseURL(doc *html.Node, fetchedURL *url.URL) *url.URL {
	var base *url.URL

	walkNode(doc, func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "base" {
			if href := attr(n, "href"); href != "" {
				if parsed, err := url.Parse(href); err == nil {
					base = fetchedURL.ResolveReference(parsed)
					return false
				}
			}
		}
		return true
	})

	if base == nil {
		return fetchedURL
	}
	return base
}

// extractTitle returns the text content of the first <title> element.
// Returns an empty string if no title is found.
func extractTitle(doc *html.Node) string {
	var title string

	walkNode(doc, func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "title" {
			title = strings.TrimSpace(textContent(n))
			return false
		}
		return true
	})

	return title
}

// countHeadings counts heading elements (h1–h6) found in the document.
func countHeadings(doc *html.Node) Headings {
	var h Headings

	walkNode(doc, func(n *html.Node) bool {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "h1":
				h.H1++
			case "h2":
				h.H2++
			case "h3":
				h.H3++
			case "h4":
				h.H4++
			case "h5":
				h.H5++
			case "h6":
				h.H6++
			}
		}
		return true
	})

	return h
}

// detectLoginForm returns true if the document contains a <form> with an
// <input type="password"> inside it.
func detectLoginForm(doc *html.Node) bool {
	found := false

	walkNode(doc, func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "form" {
			if formHasPasswordInput(n) {
				found = true
				return false
			}
		}
		return true
	})

	return found
}

// formHasPasswordInput returns true if the given form node contains
// an <input type="password"> anywhere inside it.
func formHasPasswordInput(form *html.Node) bool {
	found := false

	walkNode(form, func(n *html.Node) bool {
		if n.Type == html.ElementNode && n.Data == "input" {
			if strings.EqualFold(attr(n, "type"), "password") {
				found = true
				return false
			}
		}
		return true
	})

	return found
}

// extractLinks walks the document and returns all links found in <href>
func extractLinks(doc *html.Node, base *url.URL) []Link {
	seen := make(map[string]bool)
	var links []Link

	walkNode(doc, func(n *html.Node) bool {
		if n.Type != html.ElementNode || n.Data != "a" {
			return true
		}

		href := attr(n, "href")
		if shouldSkipHref(href) {
			return true
		}

		resolved, err := resolveURL(base, href)
		if err != nil {
			return true
		}

		// count duplicates for totals
		link := Link{
			URL:        resolved.String(),
			IsInternal: isInternal(base, resolved),
		}

		if !seen[link.URL] {
			seen[link.URL] = true
		}

		links = append(links, link)
		return true
	})

	return links
}

// shouldSkipHref returns true for hrefs that are not navigable HTTP links.
func shouldSkipHref(href string) bool {
	if href == "" || href == "#" {
		return true
	}

	lower := strings.ToLower(strings.TrimSpace(href))

	for _, prefix := range []string{"#", "javascript:", "mailto:", "tel:", "ftp:", "data:"} {
		if strings.HasPrefix(lower, prefix) {
			return true
		}
	}

	return false
}

// resolveURL resolves an href against the base URL.
func resolveURL(base *url.URL, href string) (*url.URL, error) {
	ref, err := url.Parse(href)
	if err != nil {
		return nil, err
	}
	resolved := base.ResolveReference(ref)

	// only allow http and https after resolution
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return nil, url.EscapeError("unsupported scheme")
	}

	return resolved, nil
}

// isInternal returns true if the target URL is on the same host as the base.
func isInternal(base, target *url.URL) bool {
	return normalizeHost(base.Host) == normalizeHost(target.Host)
}

// normalizeHost strips the www. prefix and lowercases the host.
func normalizeHost(host string) string {
	return strings.TrimPrefix(strings.ToLower(host), "www.")
}

// walkNode performs a depth-first traversal of the HTML node tree.
func walkNode(n *html.Node, visit func(*html.Node) bool) bool {
	if !visit(n) {
		return false
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if !walkNode(c, visit) {
			return false
		}
	}
	return true
}

// attr returns the value of the named attribute on the given node.
func attr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}
	return ""
}

// textContent returns the concatenated text content of a node and
// all its descendants.
func textContent(n *html.Node) string {
	var b strings.Builder
	walkNode(n, func(n *html.Node) bool {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		return true
	})
	return b.String()
}
