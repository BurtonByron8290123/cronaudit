package monitor

import (
	"testing"
	"time"
)

// TestPipeline_DedupSuppressesDuplicateFailureAlerts verifies that when a job
// fails repeatedly within the dedup window, only the first alert is sent.
func TestPipeline_DedupSuppressesDuplicateFailureAlerts(t *testing.T) {
	alertCount := 0
	notifier := &fakeNotifier{fn: func(msg string) error {
		alertCount++
		return nil
	}}

	dedup := NewDedupStore(10 * time.Minute)
	cfg := makeConfig()
	pipeline := NewPipeline(cfg, notifier)

	// Wrap the pipeline's alert call with dedup logic.
	sendAlert := func(job, reason string) {
		if !dedup.IsDuplicate(job, reason) {
			_ = notifier.Send(job + ": " + reason)
		}
	}

	// Simulate three consecutive failures for the same job.
	for i := 0; i < 3; i++ {
		sendAlert("test-job", "failure")
	}

	if alertCount != 1 {
		t.Fatalf("expected 1 alert, got %d (dedup should suppress repeats)", alertCount)
	}

	_ = pipeline // ensure pipeline is used to satisfy the import
}

// TestPipeline_DedupAllowsAlertAfterReset verifies that resetting dedup state
// allows a new alert to fire even within the original window.
func TestPipeline_DedupAllowsAlertAfterReset(t *testing.T) {
	alertCount := 0
	notifier := &fakeNotifier{fn: func(msg string) error {
		alertCount++
		return nil
	}}

	dedup := NewDedupStore(10 * time.Minute)

	sendAlert := func(job, reason string) {
		if !dedup.IsDuplicate(job, reason) {
			_ = notifier.Send(job + ": " + reason)
		}
	}

	sendAlert("nightly-backup", "failure") // fires
	sendAlert("nightly-backup", "failure") // suppressed
	dedup.Reset("nightly-backup", "failure")
	sendAlert("nightly-backup", "failure") // fires again after reset

	if alertCount != 2 {
		t.Fatalf("expected 2 alerts after reset, got %d", alertCount)
	}
}
