package monitor

import (
	"context"
	"fmt"
)

// RouteRule maps a set of label matchers to a specific Notifier.
type RouteRule struct {
	// RequiredLabels are key=value pairs that a job must have for this rule to match.
	RequiredLabels map[string]string
	// Notifier is the alert destination for this rule.
	Notifier Notifier
}

// AlertRouter dispatches AlertEvents to one or more Notifiers based on
// RouteRules. If no rule matches, the fallback Notifier is used (if set).
type AlertRouter struct {
	rules    []RouteRule
	fallback Notifier
}

// NewAlertRouter creates an AlertRouter with the given rules and an optional
// fallback notifier (may be nil to silently drop unmatched events).
func NewAlertRouter(rules []RouteRule, fallback Notifier) *AlertRouter {
	return &AlertRouter{
		rules:    rules,
		fallback: fallback,
	}
}

// Route sends the event to the first matching rule's Notifier. If no rule
// matches, the fallback is used. Returns an error if the chosen notifier fails.
func (r *AlertRouter) Route(ctx context.Context, event AlertEvent) error {
	for _, rule := range r.rules {
		if matchesLabels(event.Labels, rule.RequiredLabels) {
			return rule.Notifier.Send(ctx, event)
		}
	}
	if r.fallback != nil {
		return r.fallback.Send(ctx, event)
	}
	return nil
}

// Send implements Notifier so AlertRouter can be used wherever a Notifier
// is expected.
func (r *AlertRouter) Send(ctx context.Context, event AlertEvent) error {
	return r.Route(ctx, event)
}

// matchesLabels returns true when all required key/value pairs are present in
// the candidate label map.
func matchesLabels(candidate, required map[string]string) bool {
	for k, v := range required {
		if got, ok := candidate[k]; !ok || got != v {
			return false
		}
	}
	return true
}

// String returns a human-readable description of the router configuration.
func (r *AlertRouter) String() string {
	return fmt.Sprintf("AlertRouter{rules: %d, fallback: %v}", len(r.rules), r.fallback != nil)
}
