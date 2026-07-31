package transport

import (
	"net/http"

	"github.com/jljl1337/gostarter/pkg/core/transport"

	"github.com/jljl1337/gostarter/examples/full/internal/service"
)

type EndpointHandler struct {
	service         *service.EndpointService
	responseHandler *transport.ResponseHandler
}

func NewEndpointHandler(
	service *service.EndpointService,
	responseHandler *transport.ResponseHandler,
) *EndpointHandler {
	return &EndpointHandler{
		service:         service,
		responseHandler: responseHandler,
	}
}

func (h *EndpointHandler) RegisterRoutes(mux *http.ServeMux) {
	h.registerNoteRoutes(mux)
}
