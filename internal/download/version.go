package download

import (
	"fmt"
	"sync"

	httpclient "github.com/nicholas-fedor/goUpdater/internal/http"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

// versionCache provides thread-safe in-memory caching for version info to avoid redundant
// fetches within a single execution.
//
//nolint:gochecknoglobals
var versionCache sync.Map

// GetLatestVersion retrieves the latest Go version information from the official Go API.
func (d *defaultVersionFetcher) GetLatestVersion() (*httpclient.GoVersionInfo, error) {
	info, err := httpclient.GetLatestVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	return info, nil
}

// GetLatestVersionInfo fetches the latest stable Go version information from the official API.
// It returns the version info for the latest stable version or an error if not found.
// This function wraps the version package's GetLatestVersion with error handling and in-memory caching.
// The fetcher parameter allows for dependency injection, enabling proper testing with mocks.
func GetLatestVersionInfo(fetcher VersionFetcher) (*httpclient.GoVersionInfo, error) {
	if cached, ok := versionCache.Load("latest"); ok {
		if info, ok := cached.(*httpclient.GoVersionInfo); ok {
			logger.Debug("Using cached version info")

			return info, nil
		}
	}

	// Fetch from API if not cached
	info, err := fetcher.GetLatestVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest version: %w", err)
	}

	versionCache.Store("latest", info)
	logger.Debug("Cached version info for future use")

	return info, nil
}
