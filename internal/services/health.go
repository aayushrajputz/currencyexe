package services

import (
	"context"
	"exchange-rate-service/internal/models"
)

// HealthService handles health check operations
type HealthService struct {
	version string
}

// NewHealthService creates a new health service instance
func NewHealthService() *HealthService {
	return &HealthService{
		version: "1.0.0", // This could be injected from build info
	}
}

// CheckHealth performs comprehensive health checks
func (s *HealthService) CheckHealth(ctx context.Context) *models.HealthStatus {
	healthStatus := models.NewHealthStatus("ok")
	healthStatus.Version = s.version

	// Perform various health checks
	s.checkServiceHealth(healthStatus)

	return healthStatus
}

// checkServiceHealth performs internal service health checks
func (s *HealthService) checkServiceHealth(status *models.HealthStatus) {
	// Basic service health - always healthy for now
	status.AddCheck("service", "ok")

	// Add more checks here as the service grows:
	// - Database connectivity
	// - External API availability
	// - Cache connectivity
	// - Memory usage
	// - Disk space
}

// GetVersion returns the service version
func (s *HealthService) GetVersion() string {
	return s.version
}
