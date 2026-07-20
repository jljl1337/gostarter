package endpoint

import (
	"net/http"
)

func (h *EndpointHandler) registerHealthCheckRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.healthCheck)
}

func (h *EndpointHandler) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
