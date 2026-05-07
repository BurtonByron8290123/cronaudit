package monitor

import (
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	CircuitClosed CircuitState = iota // normal operation
	CircuitOpen                        // alerting suppressed
	CircuitHalfOpen                    // probing for recovery
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreakerPolicy configures thresholds for a circuit breaker.
type CircuitBreakerPolicy struct {
	FailureThreshold int           // consecutive failures before opening
	SuccessThreshold int           // consecutive successes in half-open before closing
	OpenDuration     time.Duration // how long to stay open before probing
}

type circuitEntry struct {
	state       CircuitState
	failures    int
	successes   int
	openedAt    time.Time
}

// CircuitBreakerStore tracks per-job circuit breaker state.
type CircuitBreakerStore struct {
	mu     sync.Mutex
	policy CircuitBreakerPolicy
	entries map[string]*circuitEntry
	now    func() time.Time
}

// NewCircuitBreakerStore creates a store with the given policy.
func NewCircuitBreakerStore(policy CircuitBreakerPolicy, now func() time.Time) *CircuitBreakerStore {
	if now == nil {
		now = time.Now
	}
	return &CircuitBreakerStore{
		policy:  policy,
		entries: make(map[string]*circuitEntry),
		now:     now,
	}
}

// Allow returns true if an alert should be sent for the given job key.
// It transitions state based on the outcome of the last execution.
func (c *CircuitBreakerStore) Allow(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	e := c.getOrCreate(key)
	switch e.state {
	case CircuitClosed:
		return true
	case CircuitOpen:
		if c.now().Sub(e.openedAt) >= c.policy.OpenDuration {
			e.state = CircuitHalfOpen
			e.successes = 0
			return true
		}
		return false
	case CircuitHalfOpen:
		return true
	}
	return true
}

// RecordFailure records a failure event and may open the circuit.
func (c *CircuitBreakerStore) RecordFailure(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e := c.getOrCreate(key)
	e.successes = 0
	e.failures++
	if e.state == CircuitHalfOpen || (e.state == CircuitClosed && e.failures >= c.policy.FailureThreshold) {
		e.state = CircuitOpen
		e.openedAt = c.now()
		e.failures = 0
	}
}

// RecordSuccess records a success event and may close the circuit.
func (c *CircuitBreakerStore) RecordSuccess(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e := c.getOrCreate(key)
	e.failures = 0
	if e.state == CircuitHalfOpen {
		e.successes++
		if e.successes >= c.policy.SuccessThreshold {
			e.state = CircuitClosed
			e.successes = 0
		}
	}
}

// State returns the current circuit state for a job key.
func (c *CircuitBreakerStore) State(key string) CircuitState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.getOrCreate(key).state
}

func (c *CircuitBreakerStore) getOrCreate(key string) *circuitEntry {
	if e, ok := c.entries[key]; ok {
		return e
	}
	e := &circuitEntry{state: CircuitClosed}
	c.entries[key] = e
	return e
}
