package monitor

import (
	"context"
	"errors"
	"testing"
)

// captureNotifier records the last AlertEvent it received.
type captureNotifier struct {
	last  *AlertEvent
	err   error
	calls int
}

func (c *captureNotifier) Send(_ context.Context, e AlertEvent) error {
	c.calls++
	c.last = &e
	return c.err
}

func TestAlertRouter_MatchesFirstRule(t *testing.T) {
	n1 := &captureNotifier{}
	n2 := &captureNotifier{}
	router := NewAlertRouter([]RouteRule{
		{RequiredLabels: map[string]string{"team": "ops"}, Notifier: n1},
		{RequiredLabels: map[string]string{"team": "dev"}, Notifier: n2},
	}, nil)

	event := AlertEvent{Job: "backup", Labels: map[string]string{"team": "ops"}}
	if err := router.Route(context.Background(), event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n1.calls != 1 {
		t.Errorf("expected n1 to receive 1 call, got %d", n1.calls)
	}
	if n2.calls != 0 {
		t.Errorf("expected n2 to receive 0 calls, got %d", n2.calls)
	}
}

func TestAlertRouter_FallbackWhenNoMatch(t *testing.T) {
	fb := &captureNotifier{}
	router := NewAlertRouter([]RouteRule{
		{RequiredLabels: map[string]string{"team": "ops"}, Notifier: &captureNotifier{}},
	}, fb)

	event := AlertEvent{Job: "cleanup", Labels: map[string]string{"team": "infra"}}
	if err := router.Route(context.Background(), event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.calls != 1 {
		t.Errorf("expected fallback 1 call, got %d", fb.calls)
	}
}

func TestAlertRouter_NoFallback_DropsEvent(t *testing.T) {
	router := NewAlertRouter([]RouteRule{
		{RequiredLabels: map[string]string{"team": "ops"}, Notifier: &captureNotifier{}},
	}, nil)

	event := AlertEvent{Job: "cleanup", Labels: map[string]string{"team": "infra"}}
	if err := router.Route(context.Background(), event); err != nil {
		t.Errorf("expected nil error for dropped event, got %v", err)
	}
}

func TestAlertRouter_PropagatesNotifierError(t *testing.T) {
	wantErr := errors.New("webhook down")
	n := &captureNotifier{err: wantErr}
	router := NewAlertRouter([]RouteRule{
		{RequiredLabels: map[string]string{"env": "prod"}, Notifier: n},
	}, nil)

	event := AlertEvent{Job: "sync", Labels: map[string]string{"env": "prod"}}
	if err := router.Route(context.Background(), event); !errors.Is(err, wantErr) {
		t.Errorf("expected %v, got %v", wantErr, err)
	}
}

func TestAlertRouter_SendImplementsNotifier(t *testing.T) {
	n := &captureNotifier{}
	router := NewAlertRouter([]RouteRule{
		{RequiredLabels: map[string]string{"svc": "api"}, Notifier: n},
	}, nil)

	var notifier Notifier = router // compile-time interface check
	event := AlertEvent{Job: "healthcheck", Labels: map[string]string{"svc": "api"}}
	if err := notifier.Send(context.Background(), event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.calls != 1 {
		t.Errorf("expected 1 call via Send, got %d", n.calls)
	}
}

func TestMatchesLabels_EmptyRequired(t *testing.T) {
	if !matchesLabels(map[string]string{"a": "1"}, map[string]string{}) {
		t.Error("empty required should match any candidate")
	}
}

func TestMatchesLabels_MissingKey(t *testing.T) {
	if matchesLabels(map[string]string{"a": "1"}, map[string]string{"b": "2"}) {
		t.Error("missing key should not match")
	}
}
