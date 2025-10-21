// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import "errors"

var (
	// ErrUnexpectedStatus is returned when the HTTP response status is unexpected.
	ErrUnexpectedStatus = errors.New("unexpected status")

	// ErrNoStableVersion is returned when no stable version is found in the API response.
	ErrNoStableVersion = errors.New("no stable version found")

	// ErrNetworkError is returned when a network error occurs.
	ErrNetworkError = errors.New("network error")

	// ErrCircuitBreakerOpen is returned when the circuit breaker is open.
	ErrCircuitBreakerOpen = errors.New("circuit breaker is open")

	// ErrInvalidURI is returned when the URI is invalid due to missing scheme.
	ErrInvalidURI = errors.New("invalid URI: missing scheme")

	// ErrResponseNil is returned when the response is nil.
	ErrResponseNil = errors.New("response is nil")

	// ErrStatusCodeMismatch is returned when the status code does not match the expected value.
	ErrStatusCodeMismatch = errors.New("status code mismatch")

	// ErrFailedToCreateEncoder is returned when creating a JSON encoder fails.
	ErrFailedToCreateEncoder = errors.New("failed to create JSON encoder: encoder is nil")
)
