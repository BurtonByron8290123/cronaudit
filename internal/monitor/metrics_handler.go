package monitor

import (
	"encoding/json"
	"net/http"
	"sort"
)

// MetricsHandler serves a JSON summary of per-job execution metrics.
type MetricsHandler struct {
	store *MetricsStore
}

// NewMetricsHandler creates an HTTP handler backed by the given MetricsStore.
func NewMetricsHandler(store *MetricsStore) *MetricsHandler {
	return &MetricsHandler{store: store}
}

type metricsResponse struct {
	Jobs []jobMetricsJSON `json:"jobs"`
}

type jobMetricsJSON struct {
	JobName      string  `json:"job_name"`
	TotalRuns    int64   `json:"total_runs"`
	SuccessCount int64   `json:"success_count"`
	FailureCount int64   `json:"failure_count"`
	LastDurationMs int64 `json:"last_duration_ms"`
	AvgDurationMs  int64 `json:"avg_duration_ms"`
	LastRunAt    string  `json:"last_run_at"`
}

// ServeHTTP writes a JSON metrics report for all tracked jobs.
func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	all := h.store.All()
	sort.Slice(all, func(i, j int) bool {
		return all[i].JobName < all[j].JobName
	})

	jobs := make([]jobMetricsJSON, 0, len(all))
	for _, m := range all {
		jobs = append(jobs, jobMetricsJSON{
			JobName:        m.JobName,
			TotalRuns:      m.TotalRuns,
			SuccessCount:   m.SuccessCount,
			FailureCount:   m.FailureCount,
			LastDurationMs: m.LastDuration.Milliseconds(),
			AvgDurationMs:  m.AvgDuration.Milliseconds(),
			LastRunAt:      m.LastRunAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(metricsResponse{Jobs: jobs})
}
