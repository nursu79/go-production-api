package middleware

import (
	"net/http"
	"time"
)

// Timeout wraps the request pipeline bounding it accurately utilizing http.TimeoutHandler.
func Timeout(duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// Uses the standard library HTTP Timeout handler returning a 503 natively
		return http.TimeoutHandler(next, duration, `{"error": "request timeout"}`)
	}
}
