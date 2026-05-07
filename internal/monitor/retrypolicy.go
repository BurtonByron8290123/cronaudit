package monitor

import "time"

// RetryPolicy defines how alert retries should be handled.
type RetryPolicy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// RetryState tracks retry attempts for a single alert key.
type RetryState struct {
	Attempts  int
	NextRetry time.Time
}

// RetryStore manages retry state for alert keys.
type RetryStore struct {
	policy RetryPolicy
	clock  func() time.Time
	state  map[string]*RetryState
}

// NewRetryStore creates a RetryStore with the given policy.
func NewRetryStore(policy RetryPolicy, clock func() time.Time) *RetryStore {
	if clock == nil {
		clock = time.Now
	}
	return &RetryStore{
		policy: policy,
		clock:  clock,
		state:  make(map[string]*RetryState),
	}
}

// ShouldRetry returns true if the key is eligible for a retry attempt.
func (r *RetryStore) ShouldRetry(key string) bool {
	s, ok := r.state[key]
	if !ok {
		return true
	}
	if s.Attempts >= r.policy.MaxAttempts {
		return false
	}
	return r.clock().After(s.NextRetry) || r.clock().Equal(s.NextRetry)
}

// Record records a retry attempt for the key and advances the retry schedule.
func (r *RetryStore) Record(key string) {
	s, ok := r.state[key]
	if !ok {
		s = &RetryState{}
		r.state[key] = s
	}
	s.Attempts++
	delay := r.policy.BaseDelay * (1 << (s.Attempts - 1))
	if delay > r.policy.MaxDelay {
		delay = r.policy.MaxDelay
	}
	s.NextRetry = r.clock().Add(delay)
}

// Reset clears the retry state for the given key.
func (r *RetryStore) Reset(key string) {
	delete(r.state, key)
}

// Attempts returns the current attempt count for a key.
func (r *RetryStore) Attempts(key string) int {
	if s, ok := r.state[key]; ok {
		return s.Attempts
	}
	return 0
}
