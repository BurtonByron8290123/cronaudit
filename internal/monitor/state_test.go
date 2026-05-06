package monitor

import (
	"testing"
	"time"
)

func TestJobStatus_String(t *testing.T) {
	cases := []struct {
		status   JobStatus
		expected string
	}{
		{StatusUnknown, "unknown"},
		{StatusOK, "ok"},
		{StatusFailed, "failed"},
		{StatusDrifted, "drifted"},
	}
	for _, tc := range cases {
		if got := tc.status.String(); got != tc.expected {
			t.Errorf("status %d: got %q, want %q", tc.status, got, tc.expected)
		}
	}
}

func TestStateStore_SetAndGet(t *testing.T) {
	store := NewStateStore()
	before := time.Now()
	store.Set("backup", StatusOK, "completed")
	after := time.Now()

	st := store.Get("backup")
	if st == nil {
		t.Fatal("expected state, got nil")
	}
	if st.JobName != "backup" {
		t.Errorf("got job name %q, want %q", st.JobName, "backup")
	}
	if st.Status != StatusOK {
		t.Errorf("got status %v, want StatusOK", st.Status)
	}
	if st.Message != "completed" {
		t.Errorf("got message %q, want %q", st.Message, "completed")
	}
	if st.LastSeen.Before(before) || st.LastSeen.After(after) {
		t.Errorf("LastSeen %v out of expected range", st.LastSeen)
	}
}

func TestStateStore_Get_Missing(t *testing.T) {
	store := NewStateStore()
	if st := store.Get("nonexistent"); st != nil {
		t.Errorf("expected nil for missing job, got %+v", st)
	}
}

func TestStateStore_StatusChange_UpdatesLastChange(t *testing.T) {
	store := NewStateStore()
	store.Set("sync", StatusOK, "ok")
	first := store.Get("sync")

	time.Sleep(2 * time.Millisecond)
	store.Set("sync", StatusFailed, "exit 1")
	second := store.Get("sync")

	if !second.LastChange.After(first.LastChange) {
		t.Error("expected LastChange to advance on status transition")
	}
}

func TestStateStore_SameStatus_PreservesLastChange(t *testing.T) {
	store := NewStateStore()
	store.Set("sync", StatusOK, "run 1")
	first := store.Get("sync")

	time.Sleep(2 * time.Millisecond)
	store.Set("sync", StatusOK, "run 2")
	second := store.Get("sync")

	if !second.LastChange.Equal(first.LastChange) {
		t.Error("expected LastChange to remain stable when status unchanged")
	}
}

func TestStateStore_All_ReturnsCopy(t *testing.T) {
	store := NewStateStore()
	store.Set("jobA", StatusOK, "")
	store.Set("jobB", StatusDrifted, "late")

	all := store.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}
