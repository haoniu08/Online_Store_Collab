package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the current state of the circuit breaker
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// String returns string representation of state
func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker represents a circuit breaker implementation
type CircuitBreaker struct {
	mu              sync.RWMutex
	state           State
	failureCount    int
	requestCount    int
	lastFailureTime time.Time
	successCount    int

	// Configuration
	failureThreshold int
	recoveryTimeout  time.Duration
	successThreshold int

	// Metrics
	totalRequests  int64
	totalFailures  int64
	totalSuccesses int64
}

// Config holds configuration for a circuit breaker
type Config struct {
	FailureThreshold int           // Number of failures to open circuit
	RecoveryTimeout  time.Duration // Time to wait before trying again
	SuccessThreshold int           // Number of successes to close circuit
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config Config) *CircuitBreaker {
	return &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: config.FailureThreshold,
		recoveryTimeout:  config.RecoveryTimeout,
		successThreshold: config.SuccessThreshold,
	}
}

// Execute executes the given function with circuit breaker protection
func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.totalRequests++

	// Check if circuit should transition from open to half-open
	if cb.state == StateOpen {
		if time.Since(cb.lastFailureTime) < cb.recoveryTimeout {
			return errors.New("circuit breaker is open - service temporarily unavailable")
		}
		cb.state = StateHalfOpen
		cb.successCount = 0
	}

	// Execute the function
	err := fn()

	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// recordFailure handles failure cases
func (cb *CircuitBreaker) recordFailure() {
	cb.totalFailures++
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == StateClosed {
		if cb.failureCount >= cb.failureThreshold {
			cb.state = StateOpen
		}
	} else if cb.state == StateHalfOpen {
		cb.state = StateOpen
		cb.failureCount = 1 // Reset for next attempt
	}
}

// recordSuccess handles success cases
func (cb *CircuitBreaker) recordSuccess() {
	cb.totalSuccesses++

	if cb.state == StateClosed {
		cb.failureCount = 0
	} else if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.state = StateClosed
			cb.failureCount = 0
		}
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetMetrics returns metrics about the circuit breaker
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":           cb.state.String(),
		"total_requests":  cb.totalRequests,
		"total_failures":  cb.totalFailures,
		"total_successes": cb.totalSuccesses,
		"failure_count":   cb.failureCount,
		"success_count":   cb.successCount,
		"last_failure":    cb.lastFailureTime.Format(time.RFC3339),
	}
}
