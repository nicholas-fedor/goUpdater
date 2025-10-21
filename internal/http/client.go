// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"crypto/tls"
	"net"
	"net/http"
)

// NewHTTPClient creates an optimized HTTP client for version fetching with connection pooling and timeouts.
// It configures connection pooling, keep-alive, TLS settings, and appropriate timeouts for API requests.
func NewHTTPClient() *http.Client {
	// Create optimized transport with connection pooling and keep-alive
	transport := &http.Transport{ //nolint:exhaustruct
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{ //nolint:exhaustruct
			Timeout:   DialTimeout,
			KeepAlive: KeepAlive,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          MaxIdleConns,
		IdleConnTimeout:       IdleConnTimeout,
		TLSHandshakeTimeout:   TLSHandshakeTimeout,
		ExpectContinueTimeout: ExpectContinueTimeout,
		TLSClientConfig: &tls.Config{ //nolint:exhaustruct
			MinVersion: tls.VersionTLS13,
		},
	}

	return &http.Client{
		Transport:     transport,
		CheckRedirect: nil,           // Use default redirect policy
		Jar:           nil,           // No cookie jar needed for API requests
		Timeout:       ClientTimeout, // Reasonable timeout for API requests
	}
}
