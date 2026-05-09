package monitor

import (
	"errors"
	"testing"
)

func TestTagFilter_EmptyAllowed_AllowsAll(t *testing.T) {
	f := NewTagFilter(nil)
	if !f.Allow([]string{"prod", "critical"}) {
		t.Error("expected all tags to be allowed when filter is empty")
	}
}

func TestTagFilter_EmptyAllowed_AllowsNoTags(t *testing.T) {
	f := NewTagFilter(nil)
	if !f.Allow([]string{}) {
		t.Error("expected job with no tags to be allowed when filter is empty")
	}
}

func TestTagFilter_MatchingTag_Allowed(t *testing.T) {
	f := NewTagFilter([]string{"prod", "critical"})
	if !f.Allow([]string{"staging", "critical"}) {
		t.Error("expected job with matching tag 'critical' to be allowed")
	}
}

func TestTagFilter_NoMatchingTag_Denied(t *testing.T) {
	f := NewTagFilter([]string{"prod"})
	if f.Allow([]string{"staging", "dev"}) {
		t.Error("expected job with no matching tags to be denied")
	}
}

func TestTagFilter_NoJobTags_Denied(t *testing.T) {
	f := NewTagFilter([]string{"prod"})
	if f.Allow([]string{}) {
		t.Error("expected job with no tags to be denied when filter is non-empty")
	}
}

// stubTagAlerter records the last event sent and optionally returns an error.
type stubTagAlerter struct {
	sent  []AlertEvent
	err   error
}

func (s *stubTagAlerter) Send(e AlertEvent) error {
	s.sent = append(s.sent, e)
	return s.err
}

func TestTagFilteredAlert_AllowedTag_Forwards(t *testing.T) {
	stub := &stubTagAlerter{}
	filter := NewTagFilter([]string{"prod"})
	a := NewTagFilteredAlert(stub, filter)

	event := AlertEvent{Job: "backup", Tags: []string{"prod"}}
	if err := a.Send(event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stub.sent) != 1 {
		t.Errorf("expected 1 forwarded event, got %d", len(stub.sent))
	}
}

func TestTagFilteredAlert_DeniedTag_Suppresses(t *testing.T) {
	stub := &stubTagAlerter{}
	filter := NewTagFilter([]string{"prod"})
	a := NewTagFilteredAlert(stub, filter)

	event := AlertEvent{Job: "backup", Tags: []string{"staging"}}
	if err := a.Send(event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stub.sent) != 0 {
		t.Errorf("expected 0 forwarded events, got %d", len(stub.sent))
	}
}

func TestTagFilteredAlert_PropagatesInnerError(t *testing.T) {
	want := errors.New("webhook down")
	stub := &stubTagAlerter{err: want}
	filter := NewTagFilter([]string{"prod"})
	a := NewTagFilteredAlert(stub, filter)

	event := AlertEvent{Job: "backup", Tags: []string{"prod"}}
	if err := a.Send(event); !errors.Is(err, want) {
		t.Errorf("expected error %v, got %v", want, err)
	}
}
