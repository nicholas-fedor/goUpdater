package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
)

// NewRequest creates a new HTTP request with the specified method, URL, body, headers, and context.
// It returns the created request or an error if the request creation fails.
// The headers map allows setting custom headers on the request.
// The context is used to control the request's lifecycle, such as cancellation or timeouts.
func NewRequest(ctx context.Context, method, url string,
	body io.Reader, headers map[string]string) (*http.Request, error) {
	parsedURL, err := neturl.Parse(url)
	if err != nil {
		return nil, fmt.Errorf("invalid URI: %w", err)
	}

	if parsedURL.Scheme == "" {
		return nil, ErrInvalidURI
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

// Get creates a new GET HTTP request with the specified URL, headers, and context.
// It returns the created request or an error if the request creation fails.
// This is a convenience function for GET requests without a body.
func Get(ctx context.Context, url string, headers map[string]string) (*http.Request, error) {
	return NewRequest(ctx, http.MethodGet, url, nil, headers)
}

// Post creates a new POST HTTP request with the specified URL, body, headers, and context.
// It returns the created request or an error if the request creation fails.
// The body can be any io.Reader, such as bytes.Reader or strings.Reader.
func Post(ctx context.Context, url string, body io.Reader, headers map[string]string) (*http.Request, error) {
	return NewRequest(ctx, http.MethodPost, url, body, headers)
}

// Put creates a new PUT HTTP request with the specified URL, body, headers, and context.
// It returns the created request or an error if the request creation fails.
// The body can be any io.Reader, such as bytes.Reader or strings.Reader.
func Put(ctx context.Context, url string, body io.Reader, headers map[string]string) (*http.Request, error) {
	return NewRequest(ctx, http.MethodPut, url, body, headers)
}

// Patch creates a new PATCH HTTP request with the specified URL, body, headers, and context.
// It returns the created request or an error if the request creation fails.
// The body can be any io.Reader, such as bytes.Reader or strings.Reader.
func Patch(ctx context.Context, url string, body io.Reader, headers map[string]string) (*http.Request, error) {
	return NewRequest(ctx, http.MethodPatch, url, body, headers)
}

// Delete creates a new DELETE HTTP request with the specified URL, headers, and context.
// It returns the created request or an error if the request creation fails.
// This is a convenience function for DELETE requests without a body.
func Delete(ctx context.Context, url string, headers map[string]string) (*http.Request, error) {
	return NewRequest(ctx, http.MethodDelete, url, nil, headers)
}
