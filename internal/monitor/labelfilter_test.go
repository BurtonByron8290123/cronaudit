package monitor

import (
	"errors"
	"testing"
)

func TestLabelFilter_EmptyRequired_AllowsAll(t *testing.T) {
	f := NewLabelFilter(nil)
	if !f.Allows(map[string]string{"env": "prod"}) {
		t.Fatal("empty filter should allow any labels")
	}
}

func TestLabelFilter_EmptyRequired_AllowsNoLabels(t *testing.T) {
	f := NewLabelFilter(nil)
	if !f.Allows(map[string]string{}) {
		t.Fatal("empty filter should allow empty labels")
	}
}

func TestLabelFilter_MatchingLabels_Allowed(t *testing.T) {
	f := NewLabelFilter(map[string]string{"env": "prod", "team": "ops"})
	job := map[string]string{"env": "prod", "team": "ops", "region": "us-east-1"}
	if !f.Allows(job) {
		t.Fatal("all required labels match, should be allowed")
	}
}

func TestLabelFilter_MissingLabel_Denied(t *testing.T) {
	f := NewLabelFilter(map[string]string{"env": "prod", "team": "ops"})
	job := map[string]string{"env": "prod"}
	if f.Allows(job) {
		t.Fatal("missing required label should be denied")
	}
}

func TestLabelFilter_WrongValue_Denied(t *testing.T) {
	f := NewLabelFilter(map[string]string{"env": "prod"})
	job := map[string]string{"env": "staging"}
	if f.Allows(job) {
		t.Fatal("wrong label value should be denied")
	}
}

func TestLabelFilter_NoJobLabels_Denied(t *testing.T) {
	f := NewLabelFilter(map[string]string{"env": "prod"})
	if f.Allows(map[string]string{}) {
		t.Fatal("empty job labels should be denied when labels are required")
	}
}

// --- NewLabelFilteredAlert ---

type captureNotifier struct {
	called bool
	last   AlertEvent
}

func (c *captureNotifier) Notify(e AlertEvent) error {
	c.called = true
	c.last = e
	return nil
}

func TestLabelFilteredAlert_Passes_WhenAllowed(t *testing.T) {
	cap := &captureNotifier{}
	f := NewLabelFilter(map[string]string{"env": "prod"})
	n := NewLabelFilteredAlert(cap, f)

	event := AlertEvent{JobName: "backup", Labels: map[string]string{"env": "prod"}}
	if err := n.Notify(event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cap.called {
		t.Fatal("expected inner notifier to be called")
	}
}

func TestLabelFilteredAlert_Suppresses_WhenDenied(t *testing.T) {
	cap := &captureNotifier{}
	f := NewLabelFilter(map[string]string{"env": "prod"})
	n := NewLabelFilteredAlert(cap, f)

	event := AlertEvent{JobName: "backup", Labels: map[string]string{"env": "staging"}}
	if err := n.Notify(event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.called {
		t.Fatal("expected inner notifier to be suppressed")
	}
}

func TestLabelFilteredAlert_PropagatesError(t *testing.T) {
	errNotifier := notifierFunc(func(AlertEvent) error {
		return errors.New("send failed")
	})
	f := NewLabelFilter(map[string]string{"env": "prod"})
	n := NewLabelFilteredAlert(errNotifier, f)

	event := AlertEvent{JobName: "backup", Labels: map[string]string{"env": "prod"}}
	if err := n.Notify(event); err == nil {
		t.Fatal("expected error to propagate")
	}
}
