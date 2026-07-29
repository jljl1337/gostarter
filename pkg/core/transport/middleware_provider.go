package transport

import (
	"github.com/jljl1337/gostarter/pkg/core/service"
)

// MiddlewareProvider contains all middleware functions
type MiddlewareProvider struct {
	service         *service.MiddlewareService
	responseHandler *ResponseHandler
}

// NewMiddlewareProvider creates a new middleware provider
func NewMiddlewareProvider(service *service.MiddlewareService, responseHandler *ResponseHandler) *MiddlewareProvider {
	return &MiddlewareProvider{
		service:         service,
		responseHandler: responseHandler,
	}
}

func (p *MiddlewareProvider) GetMiddlewareList() []Middleware {
	return []Middleware{
		p.Recovery(),
		p.CORS(),
		p.Logging(),
		p.Auth(),
	}
}
