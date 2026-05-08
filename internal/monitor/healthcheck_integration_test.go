package monitor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestHealthCheck_IntegrationWithPipeline verifies that a pipeline failure
// is reflected in the health endpoint response.
func TestHealthCheck_IntegrationWithPipeline(t *testing.T) {
	store := NewStateStore()
	hc := NewHealthChecker(store, 10*time.Minute)
	hc.clock = fixedHealthClock(time.Now())

	handler := NewHealthHandler(hc)

	// Simulate a job succeeding — health should be OK.
	store.Set("nightly-backup", StatusOK)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 after success, got %d", rec.Code)
	}

	// Simulate job failure — health should degrade.
	store.Set("nightly-backup", StatusFailed)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 after failure, got %d", rec.Code)
	}

	var result HealthStatus
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("invalid JSON in response: %v", err)
	}
	if result.Healthy {
		t.Error("expected Healthy=false in response body")
	}

	// Recovery — job succeeds again.
	store.Set("nightly-backup", StatusOK)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 after recovery, got %d", rec.Code)
	}
}
