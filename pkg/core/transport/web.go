package transport

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

type WebHandler struct {
	path   string
	siteFs fs.FS
}

func NewWebHandler(path string, siteFs fs.FS, subPath string) (*WebHandler, error) {
	subFs, err := fs.Sub(siteFs, subPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create sub filesystem: %w", err)
	}
	return &WebHandler{path: path, siteFs: subFs}, nil
}

func (h *WebHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc(h.path, h.serveSite)
}

func (h *WebHandler) serveSite(w http.ResponseWriter, r *http.Request) {
	// Try to serve the requested file
	filePath := strings.TrimPrefix(r.URL.Path, "/")
	if filePath == "" {
		filePath = "index.html"
	}

	// Check if file exists
	if _, err := fs.Stat(h.siteFs, filePath); err == nil {
		// File exists, serve it
		http.FileServer(http.FS(h.siteFs)).ServeHTTP(w, r)
		return
	}

	// File doesn't exist, serve index.html for SPA routing
	r.URL.Path = "/"
	http.FileServer(http.FS(h.siteFs)).ServeHTTP(w, r)
}
