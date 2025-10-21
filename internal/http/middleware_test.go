// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRetryMiddleware(t *testing.T) {
	t.Parallel()

	type args struct {
		maxRetries int
		baseDelay  time.Duration
		maxDelay   time.Duration
	}

	tests := []struct {
		name string
		args args
		want *RetryMiddleware
	}{
		{
			name: "creates retry middleware with valid parameters",
			args: args{
				maxRetries: 3,
				baseDelay:  1 * time.Second,
				maxDelay:   10 * time.Second,
			},
			want: &RetryMiddleware{
				maxRetries:        3,
				baseDelay:         1 * time.Second,
				maxDelay:          10 * time.Second,
				backoffMultiplier: 2,
			},
		},
		{
			name: "creates retry middleware with zero retries",
			args: args{
				maxRetries: 0,
				baseDelay:  500 * time.Millisecond,
				maxDelay:   5 * time.Second,
			},
			want: &RetryMiddleware{
				maxRetries:        0,
				baseDelay:         500 * time.Millisecond,
				maxDelay:          5 * time.Second,
				backoffMultiplier: 2,
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := NewRetryMiddleware(testCase.args.maxRetries, testCase.args.baseDelay, testCase.args.maxDelay)
			assert.Equal(t, testCase.want, got)
		})
	}
}

func TestRetryMiddleware_Wrap(t *testing.T) {
	t.Parallel()

	type args struct {
		next http.RoundTripper
	}

	tests := []struct {
		name string
		r    *RetryMiddleware
		args args
		want http.RoundTripper
	}{
		{
			name: "wraps round tripper with retry logic",
			r:    NewRetryMiddleware(3, 1*time.Second, 10*time.Second),
			args: args{
				next: &mockRoundTripper{},
			},
			want: &retryRoundTripper{
				next: &mockRoundTripper{},
				config: &RetryMiddleware{
					maxRetries:        3,
					baseDelay:         1 * time.Second,
					maxDelay:          10 * time.Second,
					backoffMultiplier: 2,
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.r.Wrap(testCase.args.next)
			assert.IsType(t, &retryRoundTripper{}, got)

			if rt, ok := got.(*retryRoundTripper); ok {
				assert.Equal(t, testCase.args.next, rt.next)
				assert.Equal(t, testCase.r, rt.config)
			} else {
				t.Errorf("expected *retryRoundTripper, got %T", got)
			}
		})
	}
}

func Test_retryRoundTripper_RoundTrip(t *testing.T) {
	t.Parallel()

	type args struct {
		req *http.Request
	}

	tests := []struct {
		name    string
		r       *retryRoundTripper
		args    args
		want    *http.Response
		wantErr bool
	}{
		{
			name: "successful request on first attempt",
			r: &retryRoundTripper{
				next: &mockRoundTripper{
					response: &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader("")),
					},
				},
				config: &RetryMiddleware{
					maxRetries: 3,
					baseDelay:  1 * time.Second,
					maxDelay:   10 * time.Second,
				},
			},
			args: args{
				req: &http.Request{},
			},
			want:    &http.Response{StatusCode: http.StatusOK},
			wantErr: false,
		},
		{
			name: "successful request after retries",
			r: &retryRoundTripper{
				next: &mockRoundTripper{
					response:  &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(""))},
					failCount: 2,
				},
				config: &RetryMiddleware{
					maxRetries: 3,
					baseDelay:  1 * time.Millisecond, // Use millisecond for fast test
					maxDelay:   10 * time.Millisecond,
				},
			},
			args: args{
				req: &http.Request{},
			},
			want:    &http.Response{StatusCode: http.StatusOK},
			wantErr: false,
		},
		{
			name: "fails after max retries",
			r: &retryRoundTripper{
				next: &mockRoundTripper{
					err: ErrNetworkError,
				},
				config: &RetryMiddleware{
					maxRetries: 2,
					baseDelay:  1 * time.Millisecond,
					maxDelay:   10 * time.Millisecond,
				},
			},
			args: args{
				req: &http.Request{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.r.RoundTrip(testCase.args.req)
			if got != nil {
				got.Body.Close()
			}

			if testCase.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want.StatusCode, got.StatusCode)
			}
		})
	}
}

// mockRoundTripper is a test helper that implements http.RoundTripper.
type mockRoundTripper struct {
	response  *http.Response
	err       error
	callCount int
	failCount int
}

func (m *mockRoundTripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	m.callCount++
	if m.failCount > 0 && m.callCount <= m.failCount {
		return nil, ErrNetworkError
	}

	return m.response, m.err
}

func TestNewLoggingMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want *LoggingMiddleware
	}{
		{
			name: "creates logging middleware instance",
			want: &LoggingMiddleware{},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := NewLoggingMiddleware()
			assert.Equal(t, testCase.want, got)
		})
	}
}

func TestLoggingMiddleware_Wrap(t *testing.T) {
	t.Parallel()

	type args struct {
		next http.RoundTripper
	}

	tests := []struct {
		name string
		l    *LoggingMiddleware
		args args
		want http.RoundTripper
	}{
		{
			name: "wraps round tripper with logging",
			l:    NewLoggingMiddleware(),
			args: args{
				next: &mockRoundTripper{},
			},
			want: &loggingRoundTripper{
				next: &mockRoundTripper{},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.l.Wrap(testCase.args.next)
			assert.IsType(t, &loggingRoundTripper{}, got)

			if rt, ok := got.(*loggingRoundTripper); ok {
				assert.Equal(t, testCase.args.next, rt.next)
			} else {
				t.Errorf("expected *loggingRoundTripper, got %T", got)
			}
		})
	}
}

func Test_loggingRoundTripper_RoundTrip(t *testing.T) {
	t.Parallel()

	type args struct {
		req *http.Request
	}

	tests := []struct {
		name    string
		l       *loggingRoundTripper
		args    args
		want    *http.Response
		wantErr bool
	}{
		{
			name: "logs request and successful response",
			l: &loggingRoundTripper{
				next: &mockRoundTripper{
					response: &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader("")),
					},
				},
			},
			args: args{
				req: &http.Request{Method: http.MethodGet, URL: mustParseURL("http://example.com")},
			},
			want:    &http.Response{StatusCode: http.StatusOK},
			wantErr: false,
		},
		{
			name: "logs request and error response",
			l: &loggingRoundTripper{
				next: &mockRoundTripper{err: ErrNetworkError},
			},
			args: args{
				req: &http.Request{Method: http.MethodPost, URL: mustParseURL("http://example.com/api")},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.l.RoundTrip(testCase.args.req)
			if got != nil {
				got.Body.Close()
			}

			if testCase.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want.StatusCode, got.StatusCode)
			}
		})
	}
}

// mustParseURL is a helper function for tests.
func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}

	return u
}

func TestNewTimeoutMiddleware(t *testing.T) {
	t.Parallel()

	type args struct {
		timeout time.Duration
	}

	tests := []struct {
		name string
		args args
		want *TimeoutMiddleware
	}{
		{
			name: "creates timeout middleware with specified timeout",
			args: args{
				timeout: 30 * time.Second,
			},
			want: &TimeoutMiddleware{
				timeout: 30 * time.Second,
			},
		},
		{
			name: "creates timeout middleware with zero timeout",
			args: args{
				timeout: 0,
			},
			want: &TimeoutMiddleware{
				timeout: 0,
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := NewTimeoutMiddleware(testCase.args.timeout)
			assert.Equal(t, testCase.want, got)
		})
	}
}

func TestTimeoutMiddleware_Wrap(t *testing.T) {
	t.Parallel()

	type args struct {
		next http.RoundTripper
	}

	tests := []struct {
		name string
		tr   *TimeoutMiddleware
		args args
		want http.RoundTripper
	}{
		{
			name: "wraps round tripper with timeout",
			tr:   NewTimeoutMiddleware(30 * time.Second),
			args: args{
				next: &mockRoundTripper{},
			},
			want: &timeoutRoundTripper{
				next:    &mockRoundTripper{},
				timeout: 30 * time.Second,
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.tr.Wrap(testCase.args.next)
			assert.IsType(t, &timeoutRoundTripper{}, got)

			if rt, ok := got.(*timeoutRoundTripper); ok {
				assert.Equal(t, testCase.args.next, rt.next)
				assert.Equal(t, testCase.tr.timeout, rt.timeout)
			} else {
				t.Errorf("expected *timeoutRoundTripper, got %T", got)
			}
		})
	}
}

func Test_timeoutRoundTripper_RoundTrip(t *testing.T) {
	t.Parallel()

	type args struct {
		req *http.Request
	}

	tests := []struct {
		name    string
		tr      *timeoutRoundTripper
		args    args
		want    *http.Response
		wantErr bool
	}{
		{
			name: "successful request within timeout",
			tr: &timeoutRoundTripper{
				next: &mockRoundTripper{
					response: &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader("")),
					},
				},
				timeout: 30 * time.Second,
			},
			args: args{
				req: &http.Request{},
			},
			want:    &http.Response{StatusCode: http.StatusOK},
			wantErr: false,
		},
		{
			name: "handles error from next round tripper",
			tr: &timeoutRoundTripper{
				next:    &mockRoundTripper{err: ErrNetworkError},
				timeout: 30 * time.Second,
			},
			args: args{
				req: &http.Request{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.tr.RoundTrip(testCase.args.req)
			if got != nil {
				got.Body.Close()
			}

			if testCase.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want.StatusCode, got.StatusCode)
			}
		})
	}
}

func TestNewCircuitBreakerMiddleware(t *testing.T) {
	t.Parallel()

	type args struct {
		threshold int
		timeout   time.Duration
	}

	tests := []struct {
		name string
		args args
		want *CircuitBreakerMiddleware
	}{
		{
			name: "creates circuit breaker middleware with valid parameters",
			args: args{
				threshold: 5,
				timeout:   60 * time.Second,
			},
			want: &CircuitBreakerMiddleware{
				state:     StateClosed,
				threshold: 5,
				timeout:   60 * time.Second,
				failures:  0,
			},
		},
		{
			name: "creates circuit breaker middleware with zero threshold",
			args: args{
				threshold: 0,
				timeout:   30 * time.Second,
			},
			want: &CircuitBreakerMiddleware{
				state:     StateClosed,
				threshold: 0,
				timeout:   30 * time.Second,
				failures:  0,
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := NewCircuitBreakerMiddleware(testCase.args.threshold, testCase.args.timeout)
			assert.Equal(t, testCase.want.state, got.state)
			assert.Equal(t, testCase.want.threshold, got.threshold)
			assert.Equal(t, testCase.want.timeout, got.timeout)
			assert.Equal(t, testCase.want.failures, got.failures)
			assert.NotNil(t, &got.mu)
		})
	}
}

func TestCircuitBreakerMiddleware_Wrap(t *testing.T) {
	t.Parallel()

	type args struct {
		next http.RoundTripper
	}

	tests := []struct {
		name string
		c    *CircuitBreakerMiddleware
		args args
		want http.RoundTripper
	}{
		{
			name: "wraps round tripper with circuit breaker",
			c:    NewCircuitBreakerMiddleware(5, 60*time.Second),
			args: args{
				next: &mockRoundTripper{},
			},
			want: &circuitBreakerRoundTripper{
				next: &mockRoundTripper{},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.c.Wrap(testCase.args.next)
			assert.IsType(t, &circuitBreakerRoundTripper{}, got)

			if rt, ok := got.(*circuitBreakerRoundTripper); ok {
				assert.Equal(t, testCase.args.next, rt.next)
				assert.Equal(t, testCase.c, rt.cb)
			} else {
				t.Errorf("expected *circuitBreakerRoundTripper, got %T", got)
			}
		})
	}
}

func Test_circuitBreakerRoundTripper_RoundTrip(t *testing.T) {
	t.Parallel()

	type args struct {
		req *http.Request
	}

	tests := []struct {
		name    string
		c       *circuitBreakerRoundTripper
		args    args
		want    *http.Response
		wantErr bool
	}{
		{
			name: "successful request when circuit is closed",
			c: &circuitBreakerRoundTripper{
				next: &mockRoundTripper{
					response: &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader("")),
					},
				},
				cb: NewCircuitBreakerMiddleware(3, 60*time.Second),
			},
			args: args{
				req: &http.Request{},
			},
			want:    &http.Response{StatusCode: http.StatusOK},
			wantErr: false,
		},
		{
			name: "circuit breaker opens after threshold failures",
			c: &circuitBreakerRoundTripper{
				next: &mockRoundTripper{err: ErrNetworkError},
				cb: func() *CircuitBreakerMiddleware {
					cb := NewCircuitBreakerMiddleware(2, 60*time.Second)
					cb.failures = 2 // Simulate reaching threshold

					return cb
				}(),
			},
			args: args{
				req: &http.Request{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "handles error from next round tripper",
			c: &circuitBreakerRoundTripper{
				next: &mockRoundTripper{err: ErrNetworkError},
				cb:   NewCircuitBreakerMiddleware(3, 60*time.Second),
			},
			args: args{
				req: &http.Request{},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := testCase.c.RoundTrip(testCase.args.req)
			if got != nil {
				got.Body.Close()
			}

			if testCase.wantErr {
				require.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want.StatusCode, got.StatusCode)
			}
		})
	}
}
