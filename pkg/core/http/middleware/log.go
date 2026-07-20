package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/jljl1337/gostarter/pkg/shared/env"
	"github.com/jljl1337/gostarter/pkg/shared/log"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Logging middleware logs all requests
func (m *MiddlewareProvider) Logging() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" && !env.LogHealthCheck {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			log.Infof(
				"%-6s %d %s %s %s",
				r.Method,
				wrapped.statusCode,
				r.RequestURI,
				duration,
				GetClientIP(r),
			)
		})
	}
}

// GetClientIP retrieves the client IP address from headers or connection
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (used by proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, get the first one
		ips := strings.Split(xff, ",")
		if ip := strings.TrimSpace(ips[0]); ip != "" {
			return ip
		}
	}

	// Fallback to RemoteAddr from the connection
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If no port is present, use RemoteAddr directly
		return r.RemoteAddr
	}
	return ip
}
