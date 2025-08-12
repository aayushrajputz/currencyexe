package handlers

import (
	"encoding/json"
	"net/http"

	"exchange-rate-service/internal/services"
	"exchange-rate-service/internal/utils"
)

// HealthHandler handles health check HTTP requests
type HealthHandler struct {
	healthSvc *services.HealthService
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(healthSvc *services.HealthService) *HealthHandler {
	return &HealthHandler{
		healthSvc: healthSvc,
	}
}

// CheckHealth handles GET /health requests
func (h *HealthHandler) CheckHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Perform health check
	healthStatus := h.healthSvc.CheckHealth(ctx)

	// Set appropriate HTTP status code
	statusCode := http.StatusOK
	if !healthStatus.IsHealthy() {
		statusCode = http.StatusServiceUnavailable
	}

	// send response
	utils.WriteJSON(w, statusCode, healthStatus)
}

// sendErrorResponse sends a standardized error response
func (h *HealthHandler) sendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	errorResp := map[string]string{
		"error":  message,
		"status": "error",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResp)
}
