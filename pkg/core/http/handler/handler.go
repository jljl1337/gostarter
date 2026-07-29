package handler

import "net/http"

type Handler interface {
	RegisterRoutes(*http.ServeMux)
}
