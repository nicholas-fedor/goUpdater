// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package http provides types and interfaces for HTTP operations.
package http

import (
	"net/http"
)

// Client provides HTTP client functionality for version fetching.
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

// GoVersionInfo represents version information from the Go official API.
type GoVersionInfo struct {
	Version string       `json:"version"`
	Stable  bool         `json:"stable"`
	Files   []GoFileInfo `json:"files"`
}

// GoFileInfo represents file information for a specific platform.
type GoFileInfo struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	Sha256   string `json:"sha256"`
	Size     int64  `json:"size"`
	Kind     string `json:"kind"`
}
