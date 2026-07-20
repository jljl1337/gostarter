package endpoint

import (
	"net/http"

	"github.com/jljl1337/gostarter/pkg/shared/env"
)

type versionResponse struct {
	Version string `json:"version"`
}

func (h *EndpointHandler) registerVersionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/version", h.version)
}

func (h *EndpointHandler) version(w http.ResponseWriter, r *http.Request) {
	h.responseHandler.WriteJSONResponse(w, http.StatusOK, versionResponse{Version: env.Version})
}
