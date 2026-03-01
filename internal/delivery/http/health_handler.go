package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nursu79/go-production-api/internal/domain"
)

type HealthHandler struct {
	dbPool *pgxpool.Pool
}

func NewHealthHandler(dbPool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{dbPool: dbPool}
}

func (h *HealthHandler) HealthStatus(w http.ResponseWriter, r *http.Request) {
	dbStatus := "connected"
	if h.dbPool == nil {
		dbStatus = "disconnected"
	} else if err := h.dbPool.Ping(context.Background()); err != nil {
		dbStatus = "disconnected"
	}

	health := domain.Health{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
		Database:  dbStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(health); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
