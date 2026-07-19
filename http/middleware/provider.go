package middleware

import (
	"github.com/jljl1337/gostarter/http/common"
	"github.com/jljl1337/gostarter/service"
)

// MiddlewareProvider contains all middleware functions
type MiddlewareProvider struct {
	service         *service.MiddlewareService
	responseHandler *common.ResponseHandler
}

// NewMiddlewareProvider creates a new middleware provider
func NewMiddlewareProvider(service *service.MiddlewareService, responseHandler *common.ResponseHandler) *MiddlewareProvider {
	return &MiddlewareProvider{
		service:         service,
		responseHandler: responseHandler,
	}
}
