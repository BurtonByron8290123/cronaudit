package monitor

// LabelFilter restricts alert delivery based on key-value label matching.
// A job must carry all required labels (with matching values) to pass.
type LabelFilter struct {
	required map[string]string
}

// NewLabelFilter creates a LabelFilter that requires all given key-value pairs
// to be present on a job's labels. An empty required map allows everything.
func NewLabelFilter(required map[string]string) *LabelFilter {
	copy := make(map[string]string, len(required))
	for k, v := range required {
		copy[k] = v
	}
	return &LabelFilter{required: copy}
}

// Allows returns true when every required label is present in jobLabels with
// a matching value, or when no labels are required.
func (f *LabelFilter) Allows(jobLabels map[string]string) bool {
	for k, v := range f.required {
		got, ok := jobLabels[k]
		if !ok || got != v {
			return false
		}
	}
	return true
}

// NewLabelFilteredAlert wraps a Notifier so that alerts are only forwarded
// when the event's job labels satisfy the given LabelFilter.
func NewLabelFilteredAlert(inner Notifier, filter *LabelFilter) Notifier {
	return notifierFunc(func(event AlertEvent) error {
		if !filter.Allows(event.Labels) {
			return nil
		}
		return inner.Notify(event)
	})
}

// notifierFunc is a function adapter for the Notifier interface.
type notifierFunc func(AlertEvent) error

func (fn notifierFunc) Notify(event AlertEvent) error {
	return fn(event)
}
