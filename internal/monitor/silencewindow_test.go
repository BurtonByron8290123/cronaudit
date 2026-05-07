package monitor

import (
	"testing"
	"time"
)

var epoch = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

func fixedSilenceClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestSilenceWindow_ActiveDuringWindow(t *testing.T) {
	w := SilenceWindow{Start: epoch, End: epoch.Add(1 * time.Hour)}
	if !w.Active(epoch.Add(30 * time.Minute)) {
		t.Error("expected window to be active at midpoint")
	}
}

func TestSilenceWindow_InactiveBeforeStart(t *testing.T) {
	w := SilenceWindow{Start: epoch, End: epoch.Add(1 * time.Hour)}
	if w.Active(epoch.Add(-1 * time.Minute)) {
		t.Error("expected window to be inactive before start")
	}
}

func TestSilenceWindow_InactiveAtEnd(t *testing.T) {
	w := SilenceWindow{Start: epoch, End: epoch.Add(1 * time.Hour)}
	if w.Active(epoch.Add(1 * time.Hour)) {
		t.Error("expected window to be inactive at end boundary")
	}
}

func TestSilenceStore_IsSilenced_Active(t *testing.T) {
	now := epoch.Add(30 * time.Minute)
	s := NewSilenceStore(fixedSilenceClock(now))
	s.Add("job1", epoch, epoch.Add(1*time.Hour))
	if !s.IsSilenced("job1") {
		t.Error("expected job1 to be silenced")
	}
}

func TestSilenceStore_IsSilenced_Expired(t *testing.T) {
	now := epoch.Add(2 * time.Hour)
	s := NewSilenceStore(fixedSilenceClock(now))
	s.Add("job1", epoch, epoch.Add(1*time.Hour))
	if s.IsSilenced("job1") {
		t.Error("expected job1 to not be silenced after window expires")
	}
}

func TestSilenceStore_IsSilenced_UnknownKey(t *testing.T) {
	s := NewSilenceStore(fixedSilenceClock(epoch))
	if s.IsSilenced("unknown") {
		t.Error("expected unknown job to not be silenced")
	}
}

func TestSilenceStore_MultipleWindows_OneActive(t *testing.T) {
	now := epoch.Add(30 * time.Minute)
	s := NewSilenceStore(fixedSilenceClock(now))
	s.Add("job1", epoch.Add(-2*time.Hour), epoch.Add(-1*time.Hour)) // expired
	s.Add("job1", epoch, epoch.Add(1*time.Hour))                    // active
	if !s.IsSilenced("job1") {
		t.Error("expected job1 to be silenced by second window")
	}
}

func TestSilenceStore_Purge_RemovesExpired(t *testing.T) {
	now := epoch.Add(2 * time.Hour)
	s := NewSilenceStore(fixedSilenceClock(now))
	s.Add("job1", epoch, epoch.Add(1*time.Hour))
	s.Purge()
	if s.IsSilenced("job1") {
		t.Error("expected job1 to be removed after purge")
	}
}

func TestSilenceStore_Purge_KeepsActive(t *testing.T) {
	now := epoch.Add(30 * time.Minute)
	s := NewSilenceStore(fixedSilenceClock(now))
	s.Add("job1", epoch, epoch.Add(1*time.Hour))
	s.Purge()
	if !s.IsSilenced("job1") {
		t.Error("expected job1 to remain silenced after purge of active window")
	}
}
