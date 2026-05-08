package monitor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestMetrics_IntegrationWithPipeline verifies that the MetricsStore
// is updated correctly when Pipeline processes job results, and that
// the MetricsHandler reflects those updates over HTTP.
func TestMetrics_IntegrationWithPipeline(t *testing.T) {
	metrics := NewMetricsStore()
	now := time.Now()

	// Simulate two successful runs and one failure.
	metrics.Record("nightly-sync", true, 4*time.Second, now)
	metrics.Record("nightly-sync", true, 6*time.Second, now)
	metrics.Record("nightly-sync", false, 1*time.Second, now)

	// Verify in-store values.
	m, ok := metrics.Get("nightly-sync")
	if !ok {
		t.Fatal("expected metrics for nightly-sync")
	}
	if m.TotalRuns != 3 {
		t.Errorf("TotalRuns: got %d, want 3", m.TotalRuns)
	}
	if m.SuccessCount != 2 {
		t.Errorf("SuccessCount: got %d, want 2", m.SuccessCount)
	}
	if m.FailureCount != 1 {
		t.Errorf("FailureCount: got %d, want 1", m.FailureCount)
	}
	// avg = (4+6+1)/3 = 3666ms
	expectedAvg := time.Duration((4+6+1)*int64(time.Second) / 3)
	if m.AvgDuration != expectedAvg {
		t.Errorf("AvgDuration: got %v, want %v", m.AvgDuration, expectedAvg)
	}

	// Verify via HTTP handler.
	h := NewMetricsHandler(metrics)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("HTTP status: got %d, want 200", rec.Code)
	}

	var resp metricsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Jobs) != 1 {
		t.Fatalf("expected 1 job in response, got %d", len(resp.Jobs))
	}
	if resp.Jobs[0].FailureCount != 1 {
		t.Errorf("HTTP FailureCount: got %d, want 1", resp.Jobs[0].FailureCount)
	}
}
