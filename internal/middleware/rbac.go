package middleware

import (
	"net/http"

	"github.com/nursu79/go-production-api/internal/delivery/http/response"
	"github.com/nursu79/go-production-api/internal/domain"
)

// AuthorizeRole creates a middleware that restricts access based on allowed roles.
func AuthorizeRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract role safely injected from JWTMiddleware
			role, ok := r.Context().Value(domain.UserRoleKey).(string)
			if !ok {
				response.RespondJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden: unknown role"})
				return
			}

			// Check against allowed roles
			isAllowed := false
			for _, allowedRole := range allowedRoles {
				if role == allowedRole {
					isAllowed = true
					break
				}
			}

			if !isAllowed {
				response.RespondJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden: insufficient privileges"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
