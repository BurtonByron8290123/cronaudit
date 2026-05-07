package monitor

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/yourorg/cronaudit/internal/config"
)

func TestPipeline_RetryPolicyLimitsAlerts(t *testing.T) {
	var mu sync.Mutex
	alertCount := 0

	notifier := notifierFunc(func(jobName, msg string) error {
		mu.Lock()
		alertCount++
		mu.Unlock()
		return nil
	})

	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }

	policy := RetryPolicy{
		MaxAttempts: 2,
		BaseDelay:   10 * time.Second,
		MaxDelay:    60 * time.Second,
	}
	retryStore := NewRetryStore(policy, clock)

	cfg := config.Job{Name: "backup", Command: "false", Schedule: "* * * * *"}
	history := NewHistory(10)
	state := NewStateStore()
	exec := NewExecutor(cfg, 5*time.Second)

	pipeline := NewPipeline(cfg, exec, history, state, notifier)

	ctx := context.Background()

	// First failure — should alert
	if retryStore.ShouldRetry(cfg.Name) {
		pipeline.RunJob(ctx)
		retryStore.Record(cfg.Name)
	}

	// Second failure — within delay window, should not alert
	if retryStore.ShouldRetry(cfg.Name) {
		pipeline.RunJob(ctx)
		retryStore.Record(cfg.Name)
	}

	mu.Lock()
	got := alertCount
	mu.Unlock()

	if got != 1 {
		t.Fatalf("expected 1 alert, got %d", got)
	}
}

func TestPipeline_RetryPolicyResetsOnSuccess(t *testing.T) {
	policy := RetryPolicy{MaxAttempts: 3, BaseDelay: 5 * time.Second, MaxDelay: 30 * time.Second}
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	store := NewRetryStore(policy, func() time.Time { return now })

	store.Record("backup")
	store.Record("backup")
	store.Reset("backup")

	if store.Attempts("backup") != 0 {
		t.Fatalf("expected 0 attempts after reset, got %d", store.Attempts("backup"))
	}
	if !store.ShouldRetry("backup") {
		t.Fatal("expected retry allowed after reset")
	}
}
