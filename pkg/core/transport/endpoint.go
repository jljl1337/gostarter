package transport

import (
	"net/http"

	"github.com/jljl1337/gostarter/pkg/core/service"
)

type EndpointHandler struct {
	service         *service.EndpointService
	responseHandler *ResponseHandler
	cookieGenerator *CookieGenerator
}

func NewEndpointHandler(
	service *service.EndpointService,
	responseHandler *ResponseHandler,
	cookieGenerator *CookieGenerator,
) *EndpointHandler {
	return &EndpointHandler{
		service:         service,
		responseHandler: responseHandler,
		cookieGenerator: cookieGenerator,
	}
}

func (h *EndpointHandler) RegisterRoutes(mux *http.ServeMux) {
	h.registerHealthCheckRoutes(mux)
	h.registerMetaRoutes(mux)
	h.registerAuthRoutes(mux)
	h.registerAccountRoutes(mux)
}
