package monitor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMetricsHandler_EmptyStore(t *testing.T) {
	store := NewMetricsStore()
	h := NewMetricsHandler(store)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status: got %d, want 200", rec.Code)
	}

	var resp metricsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(resp.Jobs) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(resp.Jobs))
	}
}

func TestMetricsHandler_ReturnsJobMetrics(t *testing.T) {
	store := NewMetricsStore()
	now := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)
	store.Record("backup", true, 3*time.Second, now)

	h := NewMetricsHandler(store)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	h.ServeHTTP(rec, req)

	var resp metricsResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	if len(resp.Jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(resp.Jobs))
	}
	j := resp.Jobs[0]
	if j.JobName != "backup" {
		t.Errorf("JobName: got %q, want \"backup\"", j.JobName)
	}
	if j.TotalRuns != 1 {
		t.Errorf("TotalRuns: got %d, want 1", j.TotalRuns)
	}
	if j.LastDurationMs != 3000 {
		t.Errorf("LastDurationMs: got %d, want 3000", j.LastDurationMs)
	}
}

func TestMetricsHandler_ContentTypeHeader(t *testing.T) {
	store := NewMetricsStore()
	h := NewMetricsHandler(store)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	h.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}
}

func TestMetricsHandler_JobsSortedByName(t *testing.T) {
	store := NewMetricsStore()
	now := time.Now()
	store.Record("z-job", true, time.Second, now)
	store.Record("a-job", true, time.Second, now)

	h := NewMetricsHandler(store)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	h.ServeHTTP(rec, req)

	var resp metricsResponse
	_ = json.NewDecoder(rec.Body).Decode(&resp)

	if len(resp.Jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(resp.Jobs))
	}
	if resp.Jobs[0].JobName != "a-job" {
		t.Errorf("first job: got %q, want \"a-job\"", resp.Jobs[0].JobName)
	}
}
