package transport

import (
	"fmt"
	"net/http"
)

func (m *MiddlewareProvider) Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					m.responseHandler.WriteServiceError(w, fmt.Errorf("server panic: %v", recovered))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
