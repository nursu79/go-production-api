package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// RequestLogger generates structured HTTP logs per request hooking contextual parameters perfectly.
func RequestLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			reqID := middleware.GetReqID(r.Context())

			// Only output required structured JSON preventing accidental payload leaks natively
			slog.Info("handled http request",
				slog.String("request_id", reqID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status_code", ww.Status()),
				slog.Duration("latency", time.Since(start)),
				slog.String("ip", getIP(r)),
			)
		})
	}
}
