package transport

import "net/http"

type RouteHandler interface {
	RegisterRoutes(*http.ServeMux)
}
