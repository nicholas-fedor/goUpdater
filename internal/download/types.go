// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/nicholas-fedor/goUpdater/internal/exec"
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	httpclient "github.com/nicholas-fedor/goUpdater/internal/http"
)

// defaultHTTPTimeout defines the default timeout for HTTP requests in seconds.
const defaultHTTPTimeout = 30 * time.Second

// HTTPClient provides HTTP client functionality for downloading.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// DefaultHTTPClient wraps the standard http.Client.
type DefaultHTTPClient struct {
	Client *http.Client
}

// GoVersionInfo represents the structure of a Go version from the official API.
type GoVersionInfo struct {
	Version string       `json:"version"`
	Stable  bool         `json:"stable"`
	Files   []GoFileInfo `json:"files"`
}

// GoFileInfo represents a file in a Go version for internal use.
type GoFileInfo struct {
	Filename string
	OS       string
	Arch     string
	Version  string
	Sha256   string
	Size     int
	Kind     string
}

// VersionFetcher provides version fetching functionality.
type VersionFetcher interface {
	GetLatestVersion() (*httpclient.GoVersionInfo, error)
}

// Downloader handles downloading Go archives with dependency injection for testing.
type Downloader struct {
	fs             filesystem.FileSystem
	client         HTTPClient
	executor       exec.CommandExecutor
	versionFetcher VersionFetcher
}

// NewHTTPClient creates a new HTTP client with optimized configuration for reliability and performance.
// It configures connection pooling, keep-alive, TLS settings, and HTTP/2 support for better download performance.
// The client is configured with secure TLS settings and appropriate timeouts for large file downloads.
func NewHTTPClient() *http.Client {
	// Create optimized transport with connection pooling and keep-alive
	transport := &http.Transport{ //nolint:exhaustruct
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{ //nolint:exhaustruct
			Timeout:   30 * time.Second, //nolint:mnd
			KeepAlive: 30 * time.Second, //nolint:mnd
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,              //nolint:mnd
		IdleConnTimeout:       90 * time.Second, //nolint:mnd
		TLSHandshakeTimeout:   10 * time.Second, //nolint:mnd
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{ //nolint:exhaustruct
			MinVersion: tls.VersionTLS12,
		},
	}

	return &http.Client{
		Transport:     transport,
		CheckRedirect: nil,                // Use default redirect policy
		Jar:           nil,                // No cookie jar needed for downloads
		Timeout:       defaultHTTPTimeout, // Reasonable timeout for downloads
	}
}

// Do performs the HTTP request.
// It wraps the underlying client's Do method with error handling.
func (c *DefaultHTTPClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform HTTP request: %w", err)
	}

	return resp, nil
}

// NewDownloader creates a new Downloader with the provided dependencies.
// It initializes the downloader with filesystem, HTTP client, command executor,
// and version fetcher interfaces for testing.
func NewDownloader(
	fileSystem filesystem.FileSystem,
	client HTTPClient,
	executor exec.CommandExecutor,
	versionFetcher VersionFetcher,
) *Downloader {
	return &Downloader{
		fs:             fileSystem,
		client:         client,
		executor:       executor,
		versionFetcher: versionFetcher,
	}
}
