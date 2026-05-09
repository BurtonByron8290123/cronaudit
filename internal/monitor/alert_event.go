package monitor

import "time"

// JobStatus represents the outcome of a cron job execution.
type AlertStatus string

const (
	AlertStatusFailure AlertStatus = "failure"
	AlertStatusSuccess AlertStatus = "success"
	AlertStatusDrift   AlertStatus = "drift"
)

// AlertEvent carries all context needed to dispatch an alert notification.
type AlertEvent struct {
	// Job is the name of the cron job.
	Job string `json:"job"`

	// Status describes the outcome that triggered the alert.
	Status AlertStatus `json:"status"`

	// Message is a human-readable description of the event.
	Message string `json:"message"`

	// OccurredAt is when the event was observed.
	OccurredAt time.Time `json:"occurred_at"`

	// Tags are arbitrary labels attached to the job in configuration.
	// They are used by TagFilter and routing rules.
	Tags []string `json:"tags,omitempty"`

	// DriftSeconds is populated for drift events and indicates how many
	// seconds the job ran late relative to its expected schedule.
	DriftSeconds float64 `json:"drift_seconds,omitempty"`
}

// Alerter is the interface implemented by all alert backends.
type Alerter interface {
	Send(event AlertEvent) error
}
