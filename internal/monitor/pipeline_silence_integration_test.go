package monitor

import (
	"testing"
	"time"
)

func TestPipeline_SilenceWindow_SuppressesAlert(t *testing.T) {
	cfg := makeConfig()
	alerted := false
	notifier := alertFunc(func(msg string) error {
		alerted = true
		return nil
	})

	now := epoch
	silence := NewSilenceStore(fixedSilenceClock(now))
	silence.Add(cfg.Jobs[0].Name, now.Add(-10*time.Minute), now.Add(1*time.Hour))

	p := NewPipeline(cfg, notifier)
	p.silence = silence

	err := p.RunJob(cfg.Jobs[0].Name, func() error { return fmt.Errorf("failure") })
	if err == nil {
		t.Fatal("expected job to return error")
	}
	if alerted {
		t.Error("expected alert to be suppressed during silence window")
	}
}

func TestPipeline_SilenceWindow_AllowsAlertOutsideWindow(t *testing.T) {
	cfg := makeConfig()
	alerted := false
	notifier := alertFunc(func(msg string) error {
		alerted = true
		return nil
	})

	now := epoch
	silence := NewSilenceStore(fixedSilenceClock(now))
	// window already expired
	silence.Add(cfg.Jobs[0].Name, now.Add(-2*time.Hour), now.Add(-1*time.Hour))

	p := NewPipeline(cfg, notifier)
	p.silence = silence

	err := p.RunJob(cfg.Jobs[0].Name, func() error { return fmt.Errorf("failure") })
	if err == nil {
		t.Fatal("expected job to return error")
	}
	if !alerted {
		t.Error("expected alert to fire outside silence window")
	}
}
