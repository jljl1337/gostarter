package transport

import (
	"net/http"

	"github.com/jljl1337/gostarter/pkg/shared/log"
)

func (h *EndpointHandler) registerHealthCheckRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.healthCheck)
}

func (h *EndpointHandler) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("OK")); err != nil {
		log.Errorf("Failed to write health check response: %v", err)
	}
}
