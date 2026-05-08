package monitor

import (
	"context"
	"log"
	"time"
)

// DigestSender periodically flushes a DigestStore and sends a summary alert.
type DigestSender struct {
	store    *DigestStore
	notifier Notifier
	interval time.Duration
	clock    func() time.Time
}

// Notifier is the interface for sending alert messages.
type Notifier interface {
	Send(subject, body string) error
}

// NewDigestSender creates a DigestSender that fires on the given interval.
func NewDigestSender(store *DigestStore, notifier Notifier, interval time.Duration) *DigestSender {
	return &DigestSender{
		store:    store,
		notifier: notifier,
		interval: interval,
		clock:    time.Now,
	}
}

// Run starts the periodic digest loop, blocking until ctx is cancelled.
func (ds *DigestSender) Run(ctx context.Context) {
	ticker := time.NewTicker(ds.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			ds.flush()
			return
		case <-ticker.C:
			ds.flush()
		}
	}
}

func (ds *DigestSender) flush() {
	entries := ds.store.Flush()
	if len(entries) == 0 {
		return
	}
	body := FormatDigest(entries)
	if err := ds.notifier.Send("cronaudit digest", body); err != nil {
		log.Printf("digest send error: %v", err)
	}
}
