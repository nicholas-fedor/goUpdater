// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"net/http"

	"github.com/nicholas-fedor/goUpdater/internal/exec"
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
	httpclient "github.com/nicholas-fedor/goUpdater/internal/http"
)

// HTTPClient provides HTTP client functionality for downloading.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// VersionFetcher provides version fetching functionality.
type VersionFetcher interface {
	GetLatestVersion() (*httpclient.GoVersionInfo, error)
}

// DefaultHTTPClient holds an HTTP client for dependency injection.
type DefaultHTTPClient struct {
	client *http.Client
}

// Downloader handles downloading Go archives with dependency injection for testing.
type Downloader struct {
	fs             filesystem.FileSystem
	client         HTTPClient
	executor       exec.CommandExecutor
	versionFetcher VersionFetcher
}

// defaultVersionFetcher implements VersionFetcher using the version package.
type defaultVersionFetcher struct{}
