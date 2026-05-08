package monitor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthHandler_Returns200WhenHealthy(t *testing.T) {
	store := NewStateStore()
	store.Set("job1", StatusOK)

	hc := NewHealthChecker(store, 10*time.Minute)
	hc.clock = fixedHealthClock(time.Now())

	h := NewHealthHandler(hc)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestHealthHandler_Returns503WhenUnhealthy(t *testing.T) {
	store := NewStateStore()
	store.Set("job1", StatusFailed)

	hc := NewHealthChecker(store, 10*time.Minute)
	hc.clock = fixedHealthClock(time.Now())

	h := NewHealthHandler(hc)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}
}

func TestHealthHandler_ResponseIsValidJSON(t *testing.T) {
	store := NewStateStore()
	store.Set("backup", StatusOK)

	hc := NewHealthChecker(store, 10*time.Minute)
	hc.clock = fixedHealthClock(time.Now())

	h := NewHealthHandler(hc)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	var result HealthStatus
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if _, ok := result.Jobs["backup"]; !ok {
		t.Error("expected 'backup' in jobs map")
	}
}

func TestHealthHandler_ContentTypeHeader(t *testing.T) {
	store := NewStateStore()
	hc := NewHealthChecker(store, 10*time.Minute)
	hc.clock = fixedHealthClock(time.Now())

	h := NewHealthHandler(hc)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
