// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewHTTPClient tests the NewHTTPClient function to ensure it returns a properly configured HTTP client.
func TestNewHTTPClient(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		check func(t *testing.T, client *http.Client)
	}{
		{
			name: "returns configured client with expected settings",
			check: func(t *testing.T, client *http.Client) {
				t.Helper()
				// Verify client is not nil
				assert.NotNil(t, client)

				// Verify overall client timeout
				assert.Equal(t, ClientTimeout, client.Timeout)

				// Verify transport configuration
				transport, ok := client.Transport.(*http.Transport)
				assert.True(t, ok, "Transport should be *http.Transport")

				// Verify connection pooling settings
				assert.Equal(t, MaxIdleConns, transport.MaxIdleConns)
				assert.Equal(t, IdleConnTimeout, transport.IdleConnTimeout)

				// Verify TLS settings
				assert.NotNil(t, transport.TLSClientConfig)
				assert.Equal(t, uint16(tls.VersionTLS13), transport.TLSClientConfig.MinVersion)

				// Verify other transport settings
				assert.True(t, transport.ForceAttemptHTTP2)
				assert.Equal(t, TLSHandshakeTimeout, transport.TLSHandshakeTimeout)
				assert.Equal(t, ExpectContinueTimeout, transport.ExpectContinueTimeout)

				// Verify dialer settings (nested in DialContext)
				// Note: DialContext is a function, so we can't directly inspect Dialer fields
				// but we can verify it's set
				assert.NotNil(t, transport.DialContext)

				// Verify other client settings
				assert.Nil(t, client.CheckRedirect)
				assert.Nil(t, client.Jar)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewHTTPClient()
			tt.check(t, got)
		})
	}
}
