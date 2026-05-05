package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelFailure Level = "failure"
	LevelDrift   Level = "drift"
	LevelRecovery Level = "recovery"
)

// Event holds the data for a single alert notification.
type Event struct {
	JobName   string    `json:"job_name"`
	Level     Level     `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// Notifier is the interface that wraps the Send method.
type Notifier interface {
	Send(event Event) error
}

// WebhookNotifier sends alert events to an HTTP webhook endpoint.
type WebhookNotifier struct {
	URL     string
	Timeout time.Duration
	client  *http.Client
}

// NewWebhookNotifier creates a WebhookNotifier with the given URL.
// A default timeout of 10 seconds is applied if timeout is zero.
func NewWebhookNotifier(url string, timeout time.Duration) *WebhookNotifier {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &WebhookNotifier{
		URL:     url,
		Timeout: timeout,
		client:  &http.Client{Timeout: timeout},
	}
}

// Send serialises the event as JSON and POSTs it to the webhook URL.
func (w *WebhookNotifier) Send(event Event) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("alert: marshal event: %w", err)
	}

	resp, err := w.client.Post(w.URL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("alert: post to %s: %w", w.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("alert: webhook returned non-2xx status %d", resp.StatusCode)
	}
	return nil
}
