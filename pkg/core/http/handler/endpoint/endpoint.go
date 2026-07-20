package endpoint

import (
	"net/http"

	"github.com/jljl1337/gostarter/pkg/core/http/common"
	"github.com/jljl1337/gostarter/pkg/core/service/endpoint"
)

type EndpointHandler struct {
	service         *endpoint.EndpointService
	responseHandler *common.ResponseHandler
	cookieGenerator *common.CookieGenerator
}

func NewEndpointHandler(
	service *endpoint.EndpointService,
	responseHandler *common.ResponseHandler,
	cookieGenerator *common.CookieGenerator,
) *EndpointHandler {
	return &EndpointHandler{
		service:         service,
		responseHandler: responseHandler,
		cookieGenerator: cookieGenerator,
	}
}

func (h *EndpointHandler) RegisterRoutes(mux *http.ServeMux) {
	h.registerHealthCheckRoutes(mux)
	h.registerVersionRoutes(mux)
	h.registerAuthRoutes(mux)
	h.registerAccountRoutes(mux)
}
