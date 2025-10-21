// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want Config
	}{
		{
			name: "returns default configuration",
			want: Config{
				DialTimeout:           DialTimeout,
				KeepAlive:             KeepAlive,
				MaxIdleConns:          MaxIdleConns,
				IdleConnTimeout:       IdleConnTimeout,
				TLSHandshakeTimeout:   TLSHandshakeTimeout,
				ExpectContinueTimeout: ExpectContinueTimeout,
				ClientTimeout:         ClientTimeout,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := DefaultConfig()
			assert.Equal(t, tt.want, got)
		})
	}
}
