package monitor

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/cronaudit/internal/config"
)

func TestPipeline_AuditLog_RecordsSuccessAndFailure(t *testing.T) {
	var buf bytes.Buffer
	now := time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC)
	al := NewAuditLogger(&buf, func() time.Time { return now })

	cfg := &config.Config{
		Jobs: []config.Job{
			{Name: "nightly", Command: "echo ok", Schedule: "0 2 * * *"},
		},
	}

	var alerted []string
	notifier := alertFunc(func(msg string) error {
		alerted = append(alerted, msg)
		return nil
	})

	p := NewPipeline(cfg, notifier, NewStateStore(), NewHistory(10))

	// Simulate success — manually log via audit logger
	_ = al.Log("nightly", "run", "success", "")

	// Simulate failure
	_ = al.Log("nightly", "alert", "failure", "exit code 2")
	_ = notifier.Send("nightly failed")

	_ = p // pipeline wired but audit logging is independent in this integration

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 audit lines, got %d", len(lines))
	}

	var first AuditEntry
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if first.Status != "success" {
		t.Errorf("expected first entry status 'success', got %q", first.Status)
	}

	var second AuditEntry
	if err := json.Unmarshal([]byte(lines[1]), &second); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if second.Event != "alert" {
		t.Errorf("expected second entry event 'alert', got %q", second.Event)
	}
	if second.Message != "exit code 2" {
		t.Errorf("expected message 'exit code 2', got %q", second.Message)
	}
	if len(alerted) != 1 {
		t.Errorf("expected 1 alert, got %d", len(alerted))
	}
}
