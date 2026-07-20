package endpoint

import (
	"net/http"

	"github.com/jljl1337/gostarter/pkg/shared/env"
)

type metaResponse struct {
	Version   string `json:"version"`
	CommitSHA string `json:"commit_sha"`
}

func (h *EndpointHandler) registerMetaRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/meta", h.meta)
}

func (h *EndpointHandler) meta(w http.ResponseWriter, r *http.Request) {
	h.responseHandler.WriteJSONResponse(w, http.StatusOK, metaResponse{
		Version:   env.Version,
		CommitSHA: env.CommitSHA,
	})
}
