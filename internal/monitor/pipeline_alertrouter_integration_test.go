package monitor

import (
	"context"
	"testing"
	"time"
)

// TestPipeline_AlertRouter_RoutesBasedOnLabels verifies that when the Pipeline
// uses an AlertRouter as its Notifier, failure alerts are dispatched to the
// correct downstream notifier based on job labels.
func TestPipeline_AlertRouter_RoutesBasedOnLabels(t *testing.T) {
	ops := &captureNotifier{}
	dev := &captureNotifier{}

	router := NewAlertRouter([]RouteRule{
		{RequiredLabels: map[string]string{"team": "ops"}, Notifier: ops},
		{RequiredLabels: map[string]string{"team": "dev"}, Notifier: dev},
	}, nil)

	cfg := makeConfig() // defined in pipeline_test.go
	cfg.Jobs[0].Labels = map[string]string{"team": "ops"}

	state := NewStateStore()
	history := NewHistory(10)
	pipeline := NewPipeline(cfg, router, state, history)

	// Simulate a failure for the job defined in makeConfig.
	pipeline.HandleResult(context.Background(), cfg.Jobs[0].Name, false, time.Second)

	if ops.calls != 1 {
		t.Errorf("expected ops notifier 1 call, got %d", ops.calls)
	}
	if dev.calls != 0 {
		t.Errorf("expected dev notifier 0 calls, got %d", dev.calls)
	}
}

// TestPipeline_AlertRouter_FallbackReceivesUnlabelledJob ensures that jobs
// without matching labels are forwarded to the fallback notifier.
func TestPipeline_AlertRouter_FallbackReceivesUnlabelledJob(t *testing.T) {
	ops := &captureNotifier{}
	fallback := &captureNotifier{}

	router := NewAlertRouter([]RouteRule{
		{RequiredLabels: map[string]string{"team": "ops"}, Notifier: ops},
	}, fallback)

	cfg := makeConfig()
	// Job has no labels — should fall through to fallback.
	cfg.Jobs[0].Labels = map[string]string{}

	state := NewStateStore()
	history := NewHistory(10)
	pipeline := NewPipeline(cfg, router, state, history)

	pipeline.HandleResult(context.Background(), cfg.Jobs[0].Name, false, time.Second)

	if fallback.calls != 1 {
		t.Errorf("expected fallback 1 call, got %d", fallback.calls)
	}
	if ops.calls != 0 {
		t.Errorf("expected ops 0 calls, got %d", ops.calls)
	}
}
