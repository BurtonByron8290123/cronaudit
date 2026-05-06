package monitor

import (
	"context"
	"testing"

	"github.com/cronaudit/cronaudit/internal/config"
)

func makeConfig(jobs ...config.Job) *config.Config {
	return &config.Config{Jobs: jobs}
}

func TestPipeline_RunJob_Success(t *testing.T) {
	cfg := makeConfig(config.Job{Name: "ok-job", Command: "true"})
	var alerted bool
	p := NewPipeline(cfg, func(r ExecResult) { alerted = true })

	p.RunJob(context.Background(), "ok-job")

	if alerted {
		t.Error("expected no alert for successful job")
	}
	recs := p.history.All("ok-job")
	if len(recs) != 1 {
		t.Fatalf("expected 1 history record, got %d", len(recs))
	}
	if recs[0].Status != StatusOK {
		t.Errorf("expected status OK, got %v", recs[0].Status)
	}
}

func TestPipeline_RunJob_Failure_Alerts(t *testing.T) {
	cfg := makeConfig(config.Job{Name: "bad-job", Command: "exit 1"})
	var alertedJob string
	p := NewPipeline(cfg, func(r ExecResult) { alertedJob = r.JobName })

	p.RunJob(context.Background(), "bad-job")

	if alertedJob != "bad-job" {
		t.Errorf("expected alert for bad-job, got %q", alertedJob)
	}
}

func TestPipeline_RunJob_NoAlertOnRepeatFailure(t *testing.T) {
	cfg := makeConfig(config.Job{Name: "flap-job", Command: "exit 1"})
	alertCount := 0
	p := NewPipeline(cfg, func(r ExecResult) { alertCount++ })

	p.RunJob(context.Background(), "flap-job")
	p.RunJob(context.Background(), "flap-job")

	if alertCount != 1 {
		t.Errorf("expected 1 alert (state change only), got %d", alertCount)
	}
}

func TestPipeline_RunJob_UnknownJob(t *testing.T) {
	cfg := makeConfig(config.Job{Name: "real-job", Command: "true"})
	p := NewPipeline(cfg, nil)
	// Should not panic
	p.RunJob(context.Background(), "ghost-job")

	if len(p.history.All("ghost-job")) != 0 {
		t.Error("expected no history for unknown job")
	}
}
