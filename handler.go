package gostarter

import "net/http"

type Handler interface {
	RegisterRoutes(*http.ServeMux)
}
