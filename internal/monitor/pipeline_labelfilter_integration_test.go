package monitor

import (
	"testing"
)

// TestPipeline_LabelFilter_SuppressesAlertForNonMatchingJob verifies that a
// label-filtered notifier does not fire when the failing job's labels do not
// satisfy the filter requirements.
func TestPipeline_LabelFilter_SuppressesAlertForNonMatchingJob(t *testing.T) {
	cfg, state, history, metrics := makeConfig()

	cap := &captureNotifier{}
	f := NewLabelFilter(map[string]string{"env": "prod"})
	filtered := NewLabelFilteredAlert(cap, f)

	p := NewPipeline(cfg, state, history, metrics, filtered)

	// Job labels do not match the filter.
	cfg.Jobs[0].Labels = map[string]string{"env": "staging"}

	p.RunJob(cfg.Jobs[0], false, 0)

	if cap.called {
		t.Fatal("alert should be suppressed for non-matching labels")
	}
}

// TestPipeline_LabelFilter_AllowsAlertForMatchingJob verifies that the alert
// is forwarded when the failing job's labels satisfy the filter.
func TestPipeline_LabelFilter_AllowsAlertForMatchingJob(t *testing.T) {
	cfg, state, history, metrics := makeConfig()

	cap := &captureNotifier{}
	f := NewLabelFilter(map[string]string{"env": "prod"})
	filtered := NewLabelFilteredAlert(cap, f)

	p := NewPipeline(cfg, state, history, metrics, filtered)

	// Job labels satisfy the filter.
	cfg.Jobs[0].Labels = map[string]string{"env": "prod", "team": "ops"}

	p.RunJob(cfg.Jobs[0], false, 0)

	if !cap.called {
		t.Fatal("alert should be forwarded for matching labels")
	}
}

// TestPipeline_LabelFilter_NoAlertOnSuccess confirms no alert fires on success
// regardless of label matching.
func TestPipeline_LabelFilter_NoAlertOnSuccess(t *testing.T) {
	cfg, state, history, metrics := makeConfig()

	cap := &captureNotifier{}
	f := NewLabelFilter(map[string]string{"env": "prod"})
	filtered := NewLabelFilteredAlert(cap, f)

	p := NewPipeline(cfg, state, history, metrics, filtered)
	cfg.Jobs[0].Labels = map[string]string{"env": "prod"}

	p.RunJob(cfg.Jobs[0], true, 0)

	if cap.called {
		t.Fatal("no alert expected on successful job run")
	}
}
