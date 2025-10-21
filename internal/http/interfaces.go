// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"io"
	"net/http"
)

// RequestBuilder defines an interface for constructing HTTP requests.
// Implementations should handle request creation with proper headers, body, and method.
type RequestBuilder interface {
	// BuildRequest creates an HTTP request with the specified method, URL, and optional body.
	// Returns the constructed request or an error if creation fails.
	BuildRequest(method, url string, body io.Reader) (*http.Request, error)
}

// ResponseHandler defines an interface for processing HTTP responses.
// Implementations should handle response parsing, validation, and error checking.
type ResponseHandler interface {
	// HandleResponse processes the HTTP response and returns any extracted data or error.
	// The response body should be closed by the implementation if consumed.
	HandleResponse(resp *http.Response) (interface{}, error)
}

// Middleware defines an interface for HTTP middleware that can modify or intercept requests and responses.
// Implementations should wrap the provided RoundTripper to add functionality like logging, retries, or authentication.
type Middleware interface {
	// Wrap takes a RoundTripper and returns a wrapped RoundTripper with additional behavior.
	Wrap(next http.RoundTripper) http.RoundTripper
}

// MiddlewareFunc is a function type that implements the Middleware interface.
// It allows using functions directly as middleware without defining a struct.
type MiddlewareFunc func(next http.RoundTripper) http.RoundTripper

// Wrap implements the Middleware interface for MiddlewareFunc.
func (f MiddlewareFunc) Wrap(next http.RoundTripper) http.RoundTripper {
	return f(next)
}
