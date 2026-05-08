package monitor

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// DigestEntry holds a single job event for inclusion in a digest report.
type DigestEntry struct {
	JobName   string
	Status    string
	OccuredAt time.Time
	Message   string
}

// DigestStore accumulates job events and produces periodic digest summaries.
type DigestStore struct {
	mu      sync.Mutex
	entries []DigestEntry
	clock   func() time.Time
}

// NewDigestStore creates a DigestStore using the given clock function.
func NewDigestStore(clock func() time.Time) *DigestStore {
	if clock == nil {
		clock = time.Now
	}
	return &DigestStore{clock: clock}
}

// Record appends a job event to the digest buffer.
func (d *DigestStore) Record(jobName, status, message string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.entries = append(d.entries, DigestEntry{
		JobName:   jobName,
		Status:    status,
		OccuredAt: d.clock(),
		Message:   message,
	})
}

// Flush returns all accumulated entries and clears the buffer.
func (d *DigestStore) Flush() []DigestEntry {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]DigestEntry, len(d.entries))
	copy(out, d.entries)
	d.entries = d.entries[:0]
	return out
}

// Len returns the number of buffered entries without clearing them.
func (d *DigestStore) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.entries)
}

// Format renders the digest entries as a human-readable summary string.
func FormatDigest(entries []DigestEntry) string {
	if len(entries) == 0 {
		return "No events recorded in this digest window."
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Digest: %d event(s)\n", len(entries)))
	for _, e := range entries {
		sb.WriteString(fmt.Sprintf("  [%s] %s — %s: %s\n",
			e.OccuredAt.Format(time.RFC3339), e.JobName, e.Status, e.Message))
	}
	return sb.String()
}
