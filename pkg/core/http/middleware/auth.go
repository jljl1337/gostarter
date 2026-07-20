package middleware

import (
	"context"
	"net/http"

	"github.com/jljl1337/gostarter/pkg/core/repository"
	"github.com/jljl1337/gostarter/pkg/shared/env"
)

type contextKey string

const AccountKey contextKey = "account"

func (m *MiddlewareProvider) Auth() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip for public routes
			publicRoutes := map[string]bool{
				"/version":          true,
				"/health":           true,
				"/auth/sign-up":     true,
				"/auth/pre-session": true,
				"/auth/sign-in":     true,
				"/auth/csrf-token":  true,
			}
			if publicRoutes[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			// Get session token from cookie
			cookie, err := r.Cookie(env.SessionCookieName)
			if err != nil {
				// err is not nil only if the cookie is not present
				m.responseHandler.WriteMessageResponse(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Get CSRF token from header
			CSRFToken := r.Header.Get("X-CSRF-Token")

			if CSRFToken == "" && (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete || r.Method == http.MethodPatch) {
				m.responseHandler.WriteMessageResponse(w, "CSRF token is required", http.StatusUnauthorized)
				return
			}

			// Validate session token (and CSRF token)
			account, err := m.service.GetSessionAccountAndRefreshSession(r.Context(), cookie.Value, CSRFToken)
			if err != nil {
				m.responseHandler.WriteErrorResponse(w, err)
				return
			}

			// Add account to context
			ctx := context.WithValue(r.Context(), AccountKey, account)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAccountFromContext retrieves the authenticated account from the context.
//
// It returns nil if the account is not found or is of an unexpected type.
func GetAccountFromContext(ctx context.Context) *repository.Account {
	account, ok := ctx.Value(AccountKey).(*repository.Account)
	if !ok {
		return nil
	}
	return account
}
