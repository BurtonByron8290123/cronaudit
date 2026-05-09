package monitor

import "time"

// TagFilter determines whether a job should be processed based on its tags.
// Jobs with no tags always pass. A job passes if it has at least one tag
// matching the allowed set (if any allowed tags are configured).
type TagFilter struct {
	allowed map[string]struct{}
}

// NewTagFilter creates a TagFilter that allows jobs matching any of the given
// tags. An empty allowedTags slice means all jobs are allowed.
func NewTagFilter(allowedTags []string) *TagFilter {
	allowed := make(map[string]struct{}, len(allowedTags))
	for _, t := range allowedTags {
		allowed[t] = struct{}{}
	}
	return &TagFilter{allowed: allowed}
}

// Allow returns true if the job with the given tags should be processed.
func (f *TagFilter) Allow(tags []string) bool {
	if len(f.allowed) == 0 {
		return true
	}
	for _, t := range tags {
		if _, ok := f.allowed[t]; ok {
			return true
		}
	}
	return false
}

// TagFilteredAlert wraps an Alerter and suppresses notifications for jobs
// whose tags do not match the filter.
type TagFilteredAlert struct {
	inner   Alerter
	filter  *TagFilter
	clock   func() time.Time
}

// NewTagFilteredAlert returns an Alerter that forwards to inner only when
// the job's tags pass the filter.
func NewTagFilteredAlert(inner Alerter, filter *TagFilter) *TagFilteredAlert {
	return &TagFilteredAlert{
		inner:  inner,
		filter: filter,
		clock:  time.Now,
	}
}

// Send forwards the alert only if the event's tags are allowed by the filter.
func (a *TagFilteredAlert) Send(event AlertEvent) error {
	if !a.filter.Allow(event.Tags) {
		return nil
	}
	return a.inner.Send(event)
}
