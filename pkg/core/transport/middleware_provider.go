package transport

import (
	"github.com/jljl1337/gostarter/pkg/core/service"
	"github.com/jljl1337/gostarter/pkg/shared/role"
)

// MiddlewareProvider contains all middleware functions
type MiddlewareProvider struct {
	service         *service.MiddlewareService
	roleManager     *role.RoleManager
	responseHandler *ResponseHandler
}

// NewMiddlewareProvider creates a new middleware provider
func NewMiddlewareProvider(service *service.MiddlewareService, responseHandler *ResponseHandler, roleManager *role.RoleManager) *MiddlewareProvider {
	return &MiddlewareProvider{
		service:         service,
		responseHandler: responseHandler,
		roleManager:     roleManager,
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
