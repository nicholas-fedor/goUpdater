// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

// RetryMiddleware implements retry logic with exponential backoff for HTTP requests.
// It retries failed requests up to a maximum number of attempts with increasing delays.
type RetryMiddleware struct {
	// maxRetries is the maximum number of retry attempts.
	maxRetries int
	// baseDelay is the initial delay between retries.
	baseDelay time.Duration
	// maxDelay is the maximum delay between retries.
	maxDelay time.Duration
	// backoffMultiplier is the multiplier for exponential backoff (default 2).
	backoffMultiplier float64
}

// NewRetryMiddleware creates a new RetryMiddleware with the specified configuration.
func NewRetryMiddleware(maxRetries int, baseDelay, maxDelay time.Duration) *RetryMiddleware {
	return &RetryMiddleware{
		maxRetries:        maxRetries,
		baseDelay:         baseDelay,
		maxDelay:          maxDelay,
		backoffMultiplier: 2, //nolint:mnd // exponential backoff multiplier
	}
}

// Wrap implements the Middleware interface by returning a RoundTripper that adds retry logic.
func (r *RetryMiddleware) Wrap(next http.RoundTripper) http.RoundTripper {
	return &retryRoundTripper{
		next:   next,
		config: r,
	}
}

// retryRoundTripper is the RoundTripper implementation for retry middleware.
type retryRoundTripper struct {
	next   http.RoundTripper
	config *RetryMiddleware
}

// RoundTrip executes the HTTP request with retry logic and exponential backoff.
func (r *retryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= r.config.maxRetries; attempt++ {
		resp, err := r.next.RoundTrip(req)
		if err == nil {
			return resp, nil
		}

		lastErr = err

		if attempt < r.config.maxRetries {
			delay := time.Duration(float64(r.config.baseDelay) * math.Pow(r.config.backoffMultiplier, float64(attempt)))
			if delay > r.config.maxDelay {
				delay = r.config.maxDelay
			}

			time.Sleep(delay)
		}
	}

	return nil, lastErr
}

// LoggingMiddleware implements request and response logging for HTTP operations.
// It logs outgoing requests and incoming responses using the internal logger.
type LoggingMiddleware struct{}

// NewLoggingMiddleware creates a new LoggingMiddleware instance.
func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{}
}

// Wrap implements the Middleware interface by returning a RoundTripper that adds logging.
func (l *LoggingMiddleware) Wrap(next http.RoundTripper) http.RoundTripper {
	return &loggingRoundTripper{
		next: next,
	}
}

// loggingRoundTripper is the RoundTripper implementation for logging middleware.
type loggingRoundTripper struct {
	next http.RoundTripper
}

// RoundTrip executes the HTTP request and logs both the request and response.
func (l *loggingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	logger.Infof("HTTP %s %s", req.Method, req.URL.String())

	resp, err := l.next.RoundTrip(req)
	if err != nil {
		logger.Errorf("HTTP request failed: %v", err)

		return nil, err //nolint:wrapcheck // error from retry logic
	}

	logger.Infof("HTTP response: %d", resp.StatusCode)

	return resp, nil
}

// TimeoutMiddleware implements timeout handling for HTTP requests.
// It applies a timeout to each request using context.WithTimeout.
type TimeoutMiddleware struct {
	// timeout is the duration after which the request will be cancelled.
	timeout time.Duration
}

// NewTimeoutMiddleware creates a new TimeoutMiddleware with the specified timeout duration.
func NewTimeoutMiddleware(timeout time.Duration) *TimeoutMiddleware {
	return &TimeoutMiddleware{
		timeout: timeout,
	}
}

// Wrap implements the Middleware interface by returning a RoundTripper that adds timeout handling.
func (t *TimeoutMiddleware) Wrap(next http.RoundTripper) http.RoundTripper {
	return &timeoutRoundTripper{
		next:    next,
		timeout: t.timeout,
	}
}

// timeoutRoundTripper is the RoundTripper implementation for timeout middleware.
type timeoutRoundTripper struct {
	next    http.RoundTripper
	timeout time.Duration
}

// RoundTrip executes the HTTP request with a timeout context.
func (t *timeoutRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(req.Context(), t.timeout)
	defer cancel()

	req = req.WithContext(ctx)

	return t.next.RoundTrip(req) //nolint:wrapcheck // timeout middleware wraps context
}

// CircuitBreakerMiddleware implements a circuit breaker pattern for HTTP requests.
// It prevents cascading failures by temporarily stopping requests when failures exceed a threshold.
type CircuitBreakerMiddleware struct {
	mu           sync.Mutex
	failures     int
	lastFailTime time.Time
	state        string // "closed", "open", "half-open"
	threshold    int
	timeout      time.Duration
}

// Circuit breaker states.
const (
	StateClosed   = "closed"
	StateOpen     = "open"
	StateHalfOpen = "half-open"
)

// NewCircuitBreakerMiddleware creates a new CircuitBreakerMiddleware with the specified configuration.
func NewCircuitBreakerMiddleware(threshold int, timeout time.Duration) *CircuitBreakerMiddleware {
	return &CircuitBreakerMiddleware{
		state:        StateClosed,
		threshold:    threshold,
		timeout:      timeout,
		mu:           sync.Mutex{},
		failures:     0,
		lastFailTime: time.Time{},
	}
}

// Wrap implements the Middleware interface by returning a RoundTripper that adds circuit breaker logic.
func (c *CircuitBreakerMiddleware) Wrap(next http.RoundTripper) http.RoundTripper {
	return &circuitBreakerRoundTripper{
		next: next,
		cb:   c,
	}
}

// circuitBreakerRoundTripper is the RoundTripper implementation for circuit breaker middleware.
type circuitBreakerRoundTripper struct {
	next http.RoundTripper
	cb   *CircuitBreakerMiddleware
}

// RoundTrip executes the HTTP request with circuit breaker protection.
func (c *circuitBreakerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	c.cb.mu.Lock()

	state := c.cb.state
	if state == StateOpen {
		if time.Since(c.cb.lastFailTime) > c.cb.timeout {
			c.cb.state = StateHalfOpen
			state = StateHalfOpen
		} else {
			c.cb.mu.Unlock()

			return nil, ErrCircuitBreakerOpen
		}
	}

	c.cb.mu.Unlock()

	resp, err := c.next.RoundTrip(req)

	c.cb.mu.Lock()
	defer c.cb.mu.Unlock()

	if err != nil {
		c.cb.failures++
		if c.cb.failures >= c.cb.threshold {
			c.cb.state = StateOpen
			c.cb.lastFailTime = time.Now()
		}
	} else if state == StateHalfOpen {
		c.cb.state = StateClosed
		c.cb.failures = 0
	}

	return resp, err //nolint:wrapcheck // circuit breaker middleware wraps errors
}
