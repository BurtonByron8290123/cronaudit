package monitor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// AuditEntry represents a single audit log event.
type AuditEntry struct {
	Timestamp time.Time `json:"timestamp"`
	JobName   string    `json:"job_name"`
	Event     string    `json:"event"`
	Status    string    `json:"status,omitempty"`
	Message   string    `json:"message,omitempty"`
}

// AuditLogger writes structured audit entries to an underlying writer.
type AuditLogger struct {
	mu  sync.Mutex
	out io.Writer
	now func() time.Time
}

// NewAuditLogger creates an AuditLogger writing to the given writer.
// If w is nil, os.Stdout is used.
func NewAuditLogger(w io.Writer, now func() time.Time) *AuditLogger {
	if w == nil {
		w = os.Stdout
	}
	if now == nil {
		now = time.Now
	}
	return &AuditLogger{out: w, now: now}
}

// Log writes a single audit entry as a JSON line.
func (a *AuditLogger) Log(jobName, event, status, message string) error {
	entry := AuditEntry{
		Timestamp: a.now().UTC(),
		JobName:   jobName,
		Event:     event,
		Status:    status,
		Message:   message,
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("auditlog: marshal: %w", err)
	}
	_, err = fmt.Fprintf(a.out, "%s\n", data)
	return err
}
