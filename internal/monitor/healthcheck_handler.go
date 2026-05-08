package monitor

import (
	"encoding/json"
	"net/http"
)

// HealthHandler exposes a /healthz HTTP endpoint backed by a HealthChecker.
type HealthHandler struct {
	checker *HealthChecker
}

// NewHealthHandler creates an HTTP handler for the given HealthChecker.
func NewHealthHandler(checker *HealthChecker) *HealthHandler {
	return &HealthHandler{checker: checker}
}

// ServeHTTP writes a JSON health status. Returns 200 if healthy, 503 otherwise.
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status := h.checker.Check()

	w.Header().Set("Content-Type", "application/json")
	if !status.Healthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "failed to encode health status", http.StatusInternalServerError)
	}
}
