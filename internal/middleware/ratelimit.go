package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/nursu79/go-production-api/internal/delivery/http/response"
	"github.com/nursu79/go-production-api/internal/infrastructure/redis"
	"golang.org/x/time/rate"
)

// client represents a rate limiter for a specific client IP.
type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	mu      sync.Mutex
	clients = make(map[string]*client)
)

// init runs a background goroutine to clean up old IP limiters every minute.
func init() {
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

// RateLimit creates an IP-based rate limiter middleware.
// Leverages Redis for distributed tracking across clusters. Defaults gracefully to local memory tracking if Redis crashes.
func RateLimit(r rate.Limit, burst int, redisClient *redis.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ip := getIP(req)

			// 1. If Redis is unavailable (nil), fallback to local memory tracking (Graceful Degradation)
			if redisClient == nil || redisClient.Client == nil {
				mu.Lock()
				if _, found := clients[ip]; !found {
					clients[ip] = &client{limiter: rate.NewLimiter(r, burst)}
				}
				clients[ip].lastSeen = time.Now()

				if !clients[ip].limiter.Allow() {
					mu.Unlock()
					response.RespondJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many requests"})
					return
				}
				mu.Unlock()
				next.ServeHTTP(w, req)
				return
			}

			// 2. Redis Distributed Tracking
			ctx := req.Context()
			key := "ratelimit:" + ip
			
			// Increment the request count natively
			count, err := redisClient.Client.Incr(ctx, key).Result()
			if err != nil {
				// Redis error (e.g. timeout), let it pass to keep API available
				next.ServeHTTP(w, req)
				return
			}

			// If it's the first request, set an expiry (e.g., limit X requests per 10 seconds)
			if count == 1 {
				redisClient.Client.Expire(ctx, key, 10*time.Second) 
			}

			// Apply the distributed burst threshold constraints
			if count > int64(burst) {
				response.RespondJSON(w, http.StatusTooManyRequests, map[string]string{"error": "too many requests"})
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}

// getIP attempts to extract the client IP address from the request.
func getIP(r *http.Request) string {
	// 1. Try X-Forwarded-For
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// 2. Try X-Real-IP
	xrip := r.Header.Get("X-Real-IP")
	if xrip != "" {
		return xrip
	}

	// 3. Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
