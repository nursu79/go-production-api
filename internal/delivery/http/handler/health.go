package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nursu79/go-production-api/internal/delivery/http/response"
	"github.com/nursu79/go-production-api/internal/infrastructure/redis"
)

// HealthHandler orchestrates deep health audits across the stack.
type HealthHandler struct {
	dbPool      *pgxpool.Pool
	redisClient *redis.Client
}

// NewHealthHandler instantiates the health check controller.
func NewHealthHandler(dbPool *pgxpool.Pool, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{
		dbPool:      dbPool,
		redisClient: redisClient,
	}
}

// HealthCheck performs deep pings to PostgreSQL and Redis reporting 503 if any system is degraded.
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	dbStatus := "up"
	cacheStatus := "up"
	var dbError string
	var cacheError string
	isHealthy := true

	// 1. Probe Database
	if h.dbPool == nil {
		dbStatus = "down"
		dbError = "database pool is nil"
		isHealthy = false
	} else if err := h.dbPool.Ping(ctx); err != nil {
		dbStatus = "down"
		dbError = err.Error()
		isHealthy = false
	}

	// 2. Probe Redis
	if h.redisClient == nil || h.redisClient.Client == nil {
		cacheStatus = "down"
		cacheError = "redis client is nil"
		isHealthy = false
	} else if err := h.redisClient.Client.Ping(ctx).Err(); err != nil {
		cacheStatus = "down"
		cacheError = err.Error()
		isHealthy = false
	}

	// 3. Assemble Response
	systemInfo := map[string]interface{}{
		"database": dbStatus,
		"cache":    cacheStatus,
	}
	if dbError != "" {
		systemInfo["database_error"] = dbError
	}
	if cacheError != "" {
		systemInfo["cache_error"] = cacheError
	}

	resp := map[string]interface{}{
		"status":      "available",
		"system_info": systemInfo,
	}

	if !isHealthy {
		resp["status"] = "unhealthy"
		response.RespondJSON(w, http.StatusServiceUnavailable, resp)
		return
	}

	response.RespondJSON(w, http.StatusOK, resp)
}
