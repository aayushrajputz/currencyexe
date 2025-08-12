package models

import "time"

// HealthStatus represents the health check response structure
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp,omitempty"`
	Version   string            `json:"version,omitempty"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// NewHealthStatus creates a new health status with current timestamp
func NewHealthStatus(status string) *HealthStatus {
	return &HealthStatus{
		Status:    status,
		Timestamp: time.Now().UTC(),
		Checks:    make(map[string]string),
	}
}

// AddCheck adds a health check result to the status
func (h *HealthStatus) AddCheck(name, status string) {
	if h.Checks == nil {
		h.Checks = make(map[string]string)
	}
	h.Checks[name] = status
}

// IsHealthy returns true if the overall status is "ok"
func (h *HealthStatus) IsHealthy() bool {
	return h.Status == "ok"
}

// CurrencyRate represents an exchange rate between two currencies
type CurrencyRate struct {
	From string  `json:"from"`
	To   string  `json:"to"`
	Rate float64 `json:"rate"`
	Date string  `json:"date"`
}

// ConvertResponse represents the response for currency conversion
type ConvertResponse struct {
	Amount float64 `json:"amount"`
}
