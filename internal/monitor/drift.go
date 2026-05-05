package monitor

import (
	"fmt"
	"time"
)

// DriftResult holds the result of a drift check for a single job.
type DriftResult struct {
	JobName      string
	ExpectedAt   time.Time
	ActualAt     time.Time
	Drift        time.Duration
	ExceededLimit bool
}

// String returns a human-readable summary of the drift result.
func (d DriftResult) String() string {
	return fmt.Sprintf(
		"job=%s expected=%s actual=%s drift=%s exceeded=%v",
		d.JobName,
		d.ExpectedAt.Format(time.RFC3339),
		d.ActualAt.Format(time.RFC3339),
		d.Drift.Round(time.Second),
		d.ExceededLimit,
	)
}

// CheckDrift computes the drift between an expected execution time and the
// actual execution time. It returns a DriftResult indicating whether the
// drift exceeded the provided tolerance limit.
func CheckDrift(jobName string, expected, actual time.Time, limit time.Duration) DriftResult {
	drift := actual.Sub(expected)
	if drift < 0 {
		drift = -drift
	}
	return DriftResult{
		JobName:       jobName,
		ExpectedAt:    expected,
		ActualAt:      actual,
		Drift:         drift,
		ExceededLimit: drift > limit,
	}
}
