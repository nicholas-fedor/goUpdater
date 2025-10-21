// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import "time"

// DialTimeout is the timeout for establishing network connections.
const DialTimeout = 30 * time.Second

// KeepAlive is the keep-alive duration for persistent connections.
const KeepAlive = 30 * time.Second

// MaxIdleConns is the maximum number of idle connections in the pool.
const MaxIdleConns = 100

// IdleConnTimeout is the timeout for idle connections before closing.
const IdleConnTimeout = 90 * time.Second

// TLSHandshakeTimeout is the timeout for TLS handshake completion.
const TLSHandshakeTimeout = 10 * time.Second

// ExpectContinueTimeout is the timeout for Expect-Continue header response.
const ExpectContinueTimeout = 1 * time.Second

// ClientTimeout is the overall timeout for HTTP client requests.
const ClientTimeout = 30 * time.Second

// Config holds the configuration for HTTP client settings.
type Config struct {
	// DialTimeout specifies the timeout for establishing network connections.
	DialTimeout time.Duration
	// KeepAlive specifies the keep-alive duration for persistent connections.
	KeepAlive time.Duration
	// MaxIdleConns specifies the maximum number of idle connections in the pool.
	MaxIdleConns int
	// IdleConnTimeout specifies the timeout for idle connections before closing.
	IdleConnTimeout time.Duration
	// TLSHandshakeTimeout specifies the timeout for TLS handshake completion.
	TLSHandshakeTimeout time.Duration
	// ExpectContinueTimeout specifies the timeout for Expect-Continue header response.
	ExpectContinueTimeout time.Duration
	// ClientTimeout specifies the overall timeout for HTTP client requests.
	ClientTimeout time.Duration
}

// DefaultConfig returns the default HTTP client configuration.
func DefaultConfig() Config {
	return Config{
		DialTimeout:           DialTimeout,
		KeepAlive:             KeepAlive,
		MaxIdleConns:          MaxIdleConns,
		IdleConnTimeout:       IdleConnTimeout,
		TLSHandshakeTimeout:   TLSHandshakeTimeout,
		ExpectContinueTimeout: ExpectContinueTimeout,
		ClientTimeout:         ClientTimeout,
	}
}
