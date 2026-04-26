package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aki/todo-ai/internal/models"
	"github.com/aki/todo-ai/pkg/database"
	"go.uber.org/zap"
)

type HealthHandler struct {
	db     *database.DB
	logger *zap.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *database.DB, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{db: db, logger: logger}
}

// HealthCheck handles GET /health
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connectivity
	ctx, cancel := database.NewContextWithTimeout(2 * time.Second)
	defer cancel()

	dbErr := h.db.PingContext(ctx)

	status := "healthy"
	statusCode := http.StatusOK

	if dbErr != nil {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	// Build data map
	data := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now().UTC(),
		"database":  dbErr == nil,
	}
	if dbErr != nil {
		data["database_error"] = dbErr.Error()
	}

	response := models.APIResponse{
		Success: status == "healthy",
		Data:    data,
	}

	if dbErr != nil {
		response.Message = "Service is unhealthy"
		response.Error = dbErr.Error()
	} else {
		response.Message = "Service is healthy"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// ReadinessCheck handles GET /readiness
func (h *HealthHandler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	// Simple check if service is ready to accept traffic
	response := models.APIResponse{
		Success: true,
		Message: "Service is ready",
		Data: map[string]interface{}{
			"status":    "ready",
			"timestamp": time.Now().UTC(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}