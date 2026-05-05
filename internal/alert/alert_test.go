package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/cronaudit/internal/alert"
)

func TestWebhookNotifier_Send_Success(t *testing.T) {
	var received alert.Event

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := alert.NewWebhookNotifier(server.URL, 5*time.Second)
	event := alert.Event{
		JobName: "backup-db",
		Level:   alert.LevelFailure,
		Message: "exit code 1",
	}

	if err := n.Send(event); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.JobName != "backup-db" {
		t.Errorf("job name: want backup-db, got %s", received.JobName)
	}
	if received.Level != alert.LevelFailure {
		t.Errorf("level: want failure, got %s", received.Level)
	}
	if received.Timestamp.IsZero() {
		t.Error("timestamp should be populated automatically")
	}
}

func TestWebhookNotifier_Send_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	n := alert.NewWebhookNotifier(server.URL, 5*time.Second)
	err := n.Send(alert.Event{JobName: "test", Level: alert.LevelDrift, Message: "late"})
	if err == nil {
		t.Fatal("expected error for non-2xx response, got nil")
	}
}

func TestWebhookNotifier_Send_InvalidURL(t *testing.T) {
	n := alert.NewWebhookNotifier("http://127.0.0.1:0/no-server", time.Second)
	err := n.Send(alert.Event{JobName: "test", Level: alert.LevelRecovery, Message: "ok"})
	if err == nil {
		t.Fatal("expected error for unreachable server, got nil")
	}
}
