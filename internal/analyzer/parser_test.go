package analyzer

import (
	"net/url"
	"testing"
)

func mustParse(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}

func mustParseHTML(body string) interface{} {
	doc, err := parseHTML([]byte(body))
	if err != nil {
		panic(err)
	}
	return doc
}

func TestDetectHTMLVersion(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected string
	}{
		{
			name:     "HTML5 lowercase",
			body:     `<!DOCTYPE html><html><head></head><body></body></html>`,
			expected: "HTML5",
		},
		{
			name:     "HTML5 uppercase",
			body:     `<!DOCTYPE HTML><html></html>`,
			expected: "HTML5",
		},
		{
			name:     "XHTML 1.1",
			body:     `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN">`,
			expected: "XHTML 1.1",
		},
		{
			name:     "XHTML 1.0 Strict",
			body:     `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN">`,
			expected: "XHTML 1.0 Strict",
		},
		{
			name:     "XHTML 1.0 Transitional",
			body:     `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN">`,
			expected: "XHTML 1.0 Transitional",
		},
		{
			name:     "XHTML 1.0 Frameset",
			body:     `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Frameset//EN">`,
			expected: "XHTML 1.0 Frameset",
		},
		{
			name:     "HTML 4.01 Strict",
			body:     `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Strict//EN">`,
			expected: "HTML 4.01 Strict",
		},
		{
			name:     "HTML 4.01 Transitional",
			body:     `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN">`,
			expected: "HTML 4.01 Transitional",
		},
		{
			name:     "HTML 4.01 Frameset",
			body:     `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Frameset//EN">`,
			expected: "HTML 4.01 Frameset",
		},
		{
			name:     "no DOCTYPE",
			body:     `<html><head></head><body></body></html>`,
			expected: "Unknown",
		},
		{
			name:     "unknown DOCTYPE",
			body:     `<!DOCTYPE something-custom>`,
			expected: "Unknown DOCTYPE",
		},
		{
			name:     "empty body",
			body:     ``,
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectHTMLVersion([]byte(tt.body))
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "normal title",
			html:     `<html><head><title>Hello World</title></head><body></body></html>`,
			expected: "Hello World",
		},
		{
			name:     "title with whitespace",
			html:     `<html><head><title>  Trimmed Title  </title></head></html>`,
			expected: "Trimmed Title",
		},
		{
			name:     "no title",
			html:     `<html><head></head><body></body></html>`,
			expected: "",
		},
		{
			name:     "multiple titles — first wins",
			html:     `<html><head><title>First</title><title>Second</title></head></html>`,
			expected: "First",
		},
		{
			name:     "empty title tag",
			html:     `<html><head><title></title></head></html>`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := parseHTML([]byte(tt.html))
			if err != nil {
				t.Fatalf("parseHTML failed: %v", err)
			}
			got := extractTitle(doc)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestCountHeadings(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected Headings
	}{
		{
			name: "mixed heading levels",
			html: `<html><body>
				<h1>Title</h1>
				<h2>Sub 1</h2><h2>Sub 2</h2>
				<h3>Sub-sub 1</h3><h3>Sub-sub 2</h3><h3>Sub-sub 3</h3>
				<h4>Level 4</h4>
			</body></html>`,
			expected: Headings{H1: 1, H2: 2, H3: 3, H4: 1, H5: 0, H6: 0},
		},
		{
			name:     "no headings",
			html:     `<html><body><p>no headings here</p></body></html>`,
			expected: Headings{},
		},
		{
			name:     "only h1",
			html:     `<html><body><h1>Only</h1></body></html>`,
			expected: Headings{H1: 1},
		},
		{
			name: "all levels",
			html: `<html><body>
				<h1>1</h1><h2>2</h2><h3>3</h3>
				<h4>4</h4><h5>5</h5><h6>6</h6>
			</body></html>`,
			expected: Headings{H1: 1, H2: 1, H3: 1, H4: 1, H5: 1, H6: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := parseHTML([]byte(tt.html))
			if err != nil {
				t.Fatalf("parseHTML failed: %v", err)
			}
			got := countHeadings(doc)
			if got != tt.expected {
				t.Errorf("got %+v, want %+v", got, tt.expected)
			}
		})
	}
}

func TestDetectLoginForm(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected bool
	}{
		{
			name: "form with password input",
			html: `<html><body>
				<form action="/login">
					<input type="text" name="username"/>
					<input type="password" name="password"/>
					<button>Login</button>
				</form>
			</body></html>`,
			expected: true,
		},
		{
			name:     "password input uppercase TYPE",
			html:     `<html><body><form><input TYPE="PASSWORD"/></form></body></html>`,
			expected: true,
		},
		{
			name: "form without password input",
			html: `<html><body>
				<form action="/search">
					<input type="text" name="q"/>
					<button>Search</button>
				</form>
			</body></html>`,
			expected: false,
		},
		{
			name:     "no form at all",
			html:     `<html><body><p>nothing here</p></body></html>`,
			expected: false,
		},
		{
			name:     "password input outside form — not detected",
			html:     `<html><body><input type="password"/></body></html>`,
			expected: false,
		},
		{
			name: "multiple forms — only second has password",
			html: `<html><body>
				<form><input type="text"/></form>
				<form><input type="password"/></form>
			</body></html>`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := parseHTML([]byte(tt.html))
			if err != nil {
				t.Fatalf("parseHTML failed: %v", err)
			}
			got := detectLoginForm(doc)
			if got != tt.expected {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestShouldSkipHref(t *testing.T) {
	tests := []struct {
		href     string
		expected bool
	}{
		// should skip
		{"", true},
		{"#", true},
		{"#section", true},
		{"javascript:void(0)", true},
		{"mailto:hello@example.com", true},
		{"tel:+491234567", true},
		{"ftp://example.com", true},
		{"data:text/html,<h1>hi</h1>", true},
		// should not skip
		{"https://example.com", false},
		{"http://example.com", false},
		{"/about", false},
		{"./relative", false},
		{"../parent", false},
	}

	for _, tt := range tests {
		t.Run(tt.href, func(t *testing.T) {
			got := shouldSkipHref(tt.href)
			if got != tt.expected {
				t.Errorf("shouldSkipHref(%q) = %v, want %v", tt.href, got, tt.expected)
			}
		})
	}
}

func TestIsInternal(t *testing.T) {
	base := mustParse("https://www.example.com")

	tests := []struct {
		name     string
		target   string
		expected bool
	}{
		{"same host with www", "https://www.example.com/about", true},
		{"same host without www", "https://example.com/about", true},
		{"different host entirely", "https://other.com", false},
		{"subdomain treated as external", "https://blog.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isInternal(base, mustParse(tt.target))
			if got != tt.expected {
				t.Errorf("isInternal(%q) = %v, want %v", tt.target, got, tt.expected)
			}
		})
	}
}

func TestNormalizeHost(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"www.example.com", "example.com"},
		{"example.com", "example.com"},
		{"WWW.EXAMPLE.COM", "example.com"},
		{"blog.example.com", "blog.example.com"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeHost(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeHost(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractLinks(t *testing.T) {
	base := mustParse("https://example.com")

	tests := []struct {
		name         string
		html         string
		wantTotal    int
		wantInternal int
		wantExternal int
	}{
		{
			name: "mix of internal external and skipped",
			html: `<html><body>
				<a href="/about">About</a>
				<a href="https://example.com/contact">Contact</a>
				<a href="https://external.com">External</a>
				<a href="mailto:hello@example.com">Email</a>
				<a href="#">Anchor</a>
				<a href="javascript:void(0)">JS</a>
			</body></html>`,
			wantTotal:    3,
			wantInternal: 2,
			wantExternal: 1,
		},
		{
			name: "duplicate links counted in totals",
			html: `<html><body>
				<a href="/about">About</a>
				<a href="/about">About again</a>
				<a href="https://external.com">Ext</a>
			</body></html>`,
			wantTotal:    3, // raw count includes duplicate
			wantInternal: 2,
			wantExternal: 1,
		},
		{
			name:         "no links",
			html:         `<html><body><p>no links</p></body></html>`,
			wantTotal:    0,
			wantInternal: 0,
			wantExternal: 0,
		},
		{
			name: "only skipped hrefs",
			html: `<html><body>
				<a href="#">skip</a>
				<a href="mailto:x@y.com">skip</a>
				<a href="">skip</a>
			</body></html>`,
			wantTotal:    0,
			wantInternal: 0,
			wantExternal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := parseHTML([]byte(tt.html))
			if err != nil {
				t.Fatalf("parseHTML failed: %v", err)
			}

			links := extractLinks(doc, base)

			if len(links) != tt.wantTotal {
				t.Errorf("total links = %d, want %d", len(links), tt.wantTotal)
			}

			var internal, external int
			for _, l := range links {
				if l.IsInternal {
					internal++
				} else {
					external++
				}
			}

			if internal != tt.wantInternal {
				t.Errorf("internal = %d, want %d", internal, tt.wantInternal)
			}
			if external != tt.wantExternal {
				t.Errorf("external = %d, want %d", external, tt.wantExternal)
			}
		})
	}
}

func TestExtractBaseURL(t *testing.T) {
	fetched := mustParse("https://example.com/page")

	tests := []struct {
		name         string
		html         string
		expectedHost string
	}{
		{
			name:         "with base tag overrides fetched URL",
			html:         `<html><head><base href="https://cdn.example.com/"/></head><body></body></html>`,
			expectedHost: "cdn.example.com",
		},
		{
			name:         "without base tag falls back to fetched URL",
			html:         `<html><head></head><body></body></html>`,
			expectedHost: "example.com",
		},
		{
			name:         "empty base href falls back to fetched URL",
			html:         `<html><head><base href=""/></head><body></body></html>`,
			expectedHost: "example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := parseHTML([]byte(tt.html))
			if err != nil {
				t.Fatalf("parseHTML failed: %v", err)
			}
			got := extractBaseURL(doc, fetched)
			if got.Host != tt.expectedHost {
				t.Errorf("host = %q, want %q", got.Host, tt.expectedHost)
			}
		})
	}
}
