package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// HealthHandler handles health check endpoints
type HealthHandler struct{}

// NewHealthHandler creates a new health handler
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version,omitempty"`
}

// Health returns the health status
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Ready returns the readiness status (for Kubernetes)
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	// Add dependency checks here (DB, Redis, etc.)
	resp := HealthResponse{
		Status:    "ready",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Live returns the liveness status (for Kubernetes)
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	resp := HealthResponse{
		Status:    "alive",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
