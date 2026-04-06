# web-page-analyzer

A web application that analyzes a given URL and returns information about the page.

## Building and Running

### Requirements
* [Go 1.25+](https://go.dev/doc/install)
* Make
* [Docker](https://www.docker.com/get-started/)

### Local

```bash
git clone https://github.com/conscientia/web-page-analyzer.git
cd web-page-analyzer
 
make build    # compile binary to ./bin/webanalyzer
make run      # build and run
```

Open http://localhost:8090 in your browser.

### Docker Compose

```bash
make docker-up          # build image and start (foreground)
```
Open http://localhost:8090 in your browser.

### Configuration
All configuration is via following environment variables.

| Variable           | Default | Description                                    |
|--------------------|---------|------------------------------------------------|
| PORT               | 8090    | Port the HTTP server listens on                |
| LOG_LEVEL          | info    | debug / info / warn / error                    |
| LOG_FORMAT         | text    | text (human readable) / json (production)      |
| REQUEST_TIMEOUT    | 60s     | Total time allocated for one /analyze request  |
| FETCH_TIMEOUT      | 10s     | Timeout for fetching the target page           |
| LINK_CHECK_TIMEOUT | 30s     | Total budget for all link accessibility checks |
| PER_LINK_TIMEOUT   | 3s      | Timeout per individual HEAD request            |
| MAX_WORKERS        | 15      | Concurrent HEAD requests per analysis          |

## Design Decisions and Assumptions
**Concurrency model**
1. Link accessibility checks are the only concurrent operation in the
pipeline. Everything else is synchronous DOM traversal that completes very fast.
2. For capping the number of goroutines, we have used sempahore approach - spawns
one goroutine per job, semaphore channel limits how many can run and rest queue up
waiting for a slot.
3. We did not use the fixed goroutine pool - where we prespawn N goroutines that sit idle
waiting for jobs because we know the number of links to check in advance.

**Single binary deployment**
1. The frontend is embedded into the binary via go:embed.
2. No need to copy the static index.html file.

**Works on server rendered HTML content rather Javascript rendered content**
1. Current solution only sees the raw HTML unlike the browser which sees the rendered
javascript.

## Limitations and Edge Cases

**Javascript rendered content**
1. Current solution fetches raw HTML and act directly on it.
2. Content rendered by JavaScript after page load is not visible, so single page
applications built with modern js farmework like React will not return correct results.

**No rate limiting**
1. The solution applies no rate limiting. Under heavy concurrent use, each
request fans out to MAX_WORKERS outbound HEAD checks simultaneously.

**Very large pages**
1. Pages with hundreds of links may not have all links checked within
LINK_CHECK_TIMEOUT (default 30s). Unchecked links are reported
properly and one needs to tune the timeouts for very large websites.

**Identifying link as internal or external**
1. The solution currently classifies subdomains as external, as we only
compare hosts to do classification.

## What I Would Do Differently With More Time
**Headless browser support for Javascript rendered content**
1. Launch headless browser to render javascript before doing analysis. This will
allow the solution to handle javascript heavy pages too.

**Caching**
1. Cache analysis results by URL with a configurable TTL. 
2. Repeated analysis of the same URL within the TTL would return the cached result
instantly.

**Internal vs External link classification**
1. Use effective top level domain semantics to classify subdomain as internal links.

**Background jobs**
1. Return a job ID immediately and let the client poll for results. Allows for
handling very large pages efficiently.

**Rate limiting**
1. Limit requests per IP to prevent abuse of the link checker which fans
out to many external hosts per request.

**Integration tests**
1. Add integration tests using httptest.NewServer to test the full pipeline.