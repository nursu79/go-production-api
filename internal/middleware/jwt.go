package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nursu79/go-production-api/internal/delivery/http/response"
	"github.com/nursu79/go-production-api/internal/domain"
)

// JWTMiddleware extracts and validates the Bearer token before passing requests.
func JWTMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing authorization header"})
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				response.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid authorization format"})
				return
			}

			tokenString := parts[1]

			// Parse and validate the token
			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, errors.New("unexpected signing method")
				}
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				response.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid or expired token"})
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				response.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token claims"})
				return
			}

			// Validate token purpose (prevent refresh token from acting as access token)
			if purpose, ok := claims["purpose"].(string); !ok || purpose != "access" {
				response.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token purpose"})
				return
			}

			// Extract User Subject
			sub, ok := claims["sub"].(string)
			if !ok {
				response.RespondJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing subject claim"})
				return
			}

			// Extract User Role
			role, ok := claims["role"].(string)
			if !ok {
				role = "user" // Default fallback
			}

			// Inject into Request Context
			ctx := context.WithValue(r.Context(), domain.UserIDKey, sub)
			ctx = context.WithValue(ctx, domain.UserRoleKey, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
