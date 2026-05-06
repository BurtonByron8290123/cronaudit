package monitor

import (
	"testing"
	"time"
)

func fixedDedupClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestDedupStore_FirstAlertNotDuplicate(t *testing.T) {
	store := NewDedupStore(5 * time.Minute)
	if store.IsDuplicate("backup", "failure") {
		t.Fatal("expected first alert to not be a duplicate")
	}
}

func TestDedupStore_SecondAlertWithinWindowIsDuplicate(t *testing.T) {
	now := time.Now()
	store := NewDedupStore(5 * time.Minute)
	store.now = fixedDedupClock(now)

	store.IsDuplicate("backup", "failure")
	if !store.IsDuplicate("backup", "failure") {
		t.Fatal("expected second alert within window to be a duplicate")
	}
}

func TestDedupStore_AlertAfterWindowNotDuplicate(t *testing.T) {
	now := time.Now()
	store := NewDedupStore(5 * time.Minute)
	store.now = fixedDedupClock(now)
	store.IsDuplicate("backup", "failure")

	store.now = fixedDedupClock(now.Add(6 * time.Minute))
	if store.IsDuplicate("backup", "failure") {
		t.Fatal("expected alert after window expiry to not be a duplicate")
	}
}

func TestDedupStore_IndependentKeys(t *testing.T) {
	store := NewDedupStore(5 * time.Minute)
	store.IsDuplicate("jobA", "failure")

	if store.IsDuplicate("jobB", "failure") {
		t.Fatal("different job should not be considered duplicate")
	}
	if store.IsDuplicate("jobA", "drift") {
		t.Fatal("different reason should not be considered duplicate")
	}
}

func TestDedupStore_SuppressedCount(t *testing.T) {
	now := time.Now()
	store := NewDedupStore(10 * time.Minute)
	store.now = fixedDedupClock(now)

	store.IsDuplicate("sync", "timeout") // first — not suppressed
	store.IsDuplicate("sync", "timeout") // suppressed
	store.IsDuplicate("sync", "timeout") // suppressed

	if got := store.SuppressedCount("sync", "timeout"); got != 2 {
		t.Fatalf("expected 2 suppressed, got %d", got)
	}
}

func TestDedupStore_Reset_ClearsEntry(t *testing.T) {
	now := time.Now()
	store := NewDedupStore(10 * time.Minute)
	store.now = fixedDedupClock(now)

	store.IsDuplicate("cleanup", "failure")
	store.Reset("cleanup", "failure")

	if store.IsDuplicate("cleanup", "failure") {
		t.Fatal("expected alert to not be duplicate after reset")
	}
}
