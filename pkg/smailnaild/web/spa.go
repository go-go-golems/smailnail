package web

import (
	"io"
	"io/fs"
	"net/http"
	"strings"
)

// SPAOptions configures which URL prefixes the SPA handler should not serve.
type SPAOptions struct {
	APIPrefix string // default "/api"
}

// RegisterSPA registers an SPA handler on the given mux. It serves static files
// from publicFS and falls back to index.html for client-side routing.
// API routes (matching APIPrefix) are never served by this handler.
func RegisterSPA(mux *http.ServeMux, publicFS fs.FS, opts SPAOptions) {
	if publicFS == nil {
		return
	}
	if _, err := publicFS.Open("index.html"); err != nil {
		return
	}

	apiPrefix := opts.APIPrefix
	if apiPrefix == "" {
		apiPrefix = "/api"
	}

	fileServer := http.FileServer(http.FS(publicFS))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Never serve API routes from the SPA handler
		if strings.HasPrefix(r.URL.Path, apiPrefix) {
			http.NotFound(w, r)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		f, err := publicFS.Open(path)
		if err != nil {
			// SPA fallback: serve index.html for unknown paths
			index, err := publicFS.Open("index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			defer func() { _ = index.Close() }()

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = io.Copy(w, index)
			return
		}
		_ = f.Close()
		fileServer.ServeHTTP(w, r)
	}))
}
