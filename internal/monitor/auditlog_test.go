package monitor

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

var fixedAuditTime = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)

func fixedAuditClock() time.Time { return fixedAuditTime }

func TestAuditLogger_WritesJSONLine(t *testing.T) {
	var buf bytes.Buffer
	al := NewAuditLogger(&buf, fixedAuditClock)

	if err := al.Log("backup", "run", "success", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := strings.TrimSpace(buf.String())
	var entry AuditEntry
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if entry.JobName != "backup" {
		t.Errorf("expected job_name 'backup', got %q", entry.JobName)
	}
	if entry.Event != "run" {
		t.Errorf("expected event 'run', got %q", entry.Event)
	}
	if entry.Status != "success" {
		t.Errorf("expected status 'success', got %q", entry.Status)
	}
	if !entry.Timestamp.Equal(fixedAuditTime) {
		t.Errorf("unexpected timestamp: %v", entry.Timestamp)
	}
}

func TestAuditLogger_MultipleEntries(t *testing.T) {
	var buf bytes.Buffer
	al := NewAuditLogger(&buf, fixedAuditClock)

	_ = al.Log("job1", "run", "success", "")
	_ = al.Log("job2", "alert", "failure", "exit code 1")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	var e2 AuditEntry
	if err := json.Unmarshal([]byte(lines[1]), &e2); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if e2.Message != "exit code 1" {
		t.Errorf("expected message 'exit code 1', got %q", e2.Message)
	}
}

func TestAuditLogger_DefaultsToStdout(t *testing.T) {
	al := NewAuditLogger(nil, fixedAuditClock)
	if al.out == nil {
		t.Error("expected non-nil writer")
	}
}

func TestAuditLogger_DefaultClock(t *testing.T) {
	var buf bytes.Buffer
	al := NewAuditLogger(&buf, nil)
	before := time.Now().UTC()
	_ = al.Log("job", "run", "ok", "")
	after := time.Now().UTC()

	var entry AuditEntry
	_ = json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &entry)
	if entry.Timestamp.Before(before) || entry.Timestamp.After(after) {
		t.Errorf("timestamp %v not in expected range [%v, %v]", entry.Timestamp, before, after)
	}
}
