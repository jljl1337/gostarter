package transport

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/jljl1337/gostarter/pkg/core/repository"
	"github.com/jljl1337/gostarter/pkg/shared/env"
	"github.com/jljl1337/gostarter/pkg/shared/role"
)

type contextKey string

const AccountKey contextKey = "account"

func (m *MiddlewareProvider) Auth() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip for public routes
			publicRoutes := map[string]bool{
				"/meta":             true,
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
				m.responseHandler.WriteMessage(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Get CSRF token from header
			CSRFToken := r.Header.Get("X-CSRF-Token")

			if CSRFToken == "" && (r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodDelete || r.Method == http.MethodPatch) {
				m.responseHandler.WriteMessage(w, "CSRF token is required", http.StatusUnauthorized)
				return
			}

			// Validate session token (and CSRF token)
			account, err := m.service.GetSessionAccountAndRefreshSession(r.Context(), cookie.Value, CSRFToken)
			if err != nil {
				m.responseHandler.WriteServiceError(w, err)
				return
			}

			// Authorize subpath based on account role
			if m.roleManager.HasMultipleRoles() {
				authorized, err := authorizeSubpath(m.roleManager, account.Role, r.URL.Path)
				if err != nil {
					m.responseHandler.WriteErrorf(w, "failed to authorize subpath: %v", err)
					return
				}

				if !authorized {
					// TODO: add a default not found handler
					// Pretend this route does not exist to avoid revealing
					// information about the existence of the route
					w.WriteHeader(http.StatusNotFound)
					return
				}
			}

			// Add account to context
			ctx := context.WithValue(r.Context(), AccountKey, account)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func authorizeSubpath(roleManager *role.RoleManager, accountRole, path string) (bool, error) {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	// Allow if there are no segments
	if len(segments) == 0 {
		return true, nil
	}

	firstSegment := segments[0]
	// Allow if the first segment is not a role
	if !roleManager.HasRole(firstSegment) {
		return true, nil
	}

	compare, err := roleManager.CompareRoles(accountRole, firstSegment)
	if err != nil {
		return false, fmt.Errorf("failed to compare roles: %v", err)
	}

	if compare < 0 {
		return false, nil
	}

	return true, nil
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
