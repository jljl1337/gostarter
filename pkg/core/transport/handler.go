package transport

import "net/http"

type Handler interface {
	RegisterRoutes(*http.ServeMux)
}
