// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

// GetLatestVersion fetches the latest stable Go version information from the official API.
// It returns the version info for the current platform or an error if not found.
func GetLatestVersion() (*GoVersionInfo, error) {
	client := NewHTTPClient()

	return getLatestVersionWithClient(client)
}

// getLatestVersionWithClient fetches the latest stable Go version from the official API using the provided HTTP client.
// It returns the version info for the current platform or an error if not found.
// This function enables dependency injection for testing purposes.
func getLatestVersionWithClient(client Client) (*GoVersionInfo, error) {
	logger.Debug("Fetching latest Go version information from official API")

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://go.dev/dl/?mode=json", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set proper headers for API requests
	req.Header.Set("User-Agent", "goUpdater/dev")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch version info: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d: %w", resp.StatusCode, ErrUnexpectedStatus)
	}

	var versions []GoVersionInfo

	err = json.NewDecoder(resp.Body).Decode(&versions)
	if err != nil {
		return nil, fmt.Errorf("failed to decode version info: %w", err)
	}

	logger.Debugf("Raw API response decoded: %+v", versions)

	// Find the latest stable version
	for _, v := range versions {
		if v.Stable {
			logger.Debugf("Found stable version: %s", v.Version)

			return &v, nil
		}
	}

	return nil, ErrNoStableVersion
}
