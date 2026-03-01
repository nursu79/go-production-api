package middleware

import (
	"net/http"
)

// SecureHeaders sets rigid REST API constraints utilizing X-Frame and CSP protections preventing browser-based snooping patterns globally.
func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")

		next.ServeHTTP(w, r)
	})
}

// CORS specifies Cross-Origin rules. Maps heavily depending on active environment variables isolating specific domains cleanly inside Production.
func CORS(env string, allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Dynamically allow wildcard for development
			if env == "development" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				// Production: Validate explicitly
				allowed := false
				for _, o := range allowedOrigins {
					if origin == o {
						allowed = true
						break
					}
				}
				if allowed {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Intercept OPTONS logic safely escaping early returning a valid 204 StatusNoContent mechanism caching browser checks.
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Max-Age", "86400") // Cache properties for 24h
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
