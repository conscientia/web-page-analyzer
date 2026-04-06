package server

import "github.com/conscientia/web-page-analyzer/web"

// indexHTML holds the embedded frontend served at GET /.
// Baked into the binary at compile time via go:embed in web/web.go.
var indexHTML = web.IndexHTML
