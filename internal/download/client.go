// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	httpclient "github.com/nicholas-fedor/goUpdater/internal/http"
	"github.com/nicholas-fedor/goUpdater/internal/logger"
	"github.com/nicholas-fedor/goUpdater/internal/version"
)

// retryDelay defines the base delay between retry attempts.
const retryDelay = 2 * time.Second

// maxRetries defines the maximum number of retry attempts for failed downloads.
const maxRetries = 3

// validateDownloadURL validates the URL for security and correctness before creating a download request.
// It ensures the URL is well-formed, uses HTTPS, and doesn't contain dangerous schemes or paths.
//

func validateDownloadURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return ErrEmptyURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidURL, err)
	}

	// Only allow HTTPS scheme for security
	if parsedURL.Scheme != "https" {
		return fmt.Errorf("%w: %s", ErrInvalidURLScheme, parsedURL.Scheme)
	}

	// Ensure host is present and valid
	if parsedURL.Host == "" {
		return ErrInvalidURLHost
	}

	// Prevent localhost and private IP ranges for security
	hostname := parsedURL.Hostname()
	if strings.Contains(strings.ToLower(hostname), "localhost") {
		return ErrInvalidURLHost
	}

	if ip := net.ParseIP(hostname); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() {
			return ErrInvalidURLHost
		}
	}

	// Basic path validation - prevent directory traversal
	if strings.Contains(parsedURL.Path, "..") {
		return ErrDirectoryTraversal
	}

	return nil
}

// createDownloadRequest creates an HTTP GET request for the given URL with the provided context and proper headers.
// It sets appropriate headers for modern HTTP clients to ensure compatibility and proper content negotiation.
func (d *Downloader) createDownloadRequest(ctx context.Context, url string) (*http.Request, error) {
	logger.Debugf("Creating HTTP request for: %s", url)

	// Validate URL for security
	err := validateDownloadURL(url)
	if err != nil {
		return nil, fmt.Errorf("URL validation failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set proper headers for modern HTTP clients
	req.Header.Set("User-Agent", "goUpdater/"+version.GetClientVersion())
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	return req, nil
}

// validateSecurityHeaders validates security-related HTTP headers in the response.
// It checks for the presence of important security headers and logs warnings if they're missing.
func validateSecurityHeaders(resp *http.Response) error {
	var warnings []string

	// Check for Content-Security-Policy
	if csp := resp.Header.Get("Content-Security-Policy"); csp == "" {
		warnings = append(warnings, "missing Content-Security-Policy header")
	}

	// Check for X-Content-Type-Options
	if cto := resp.Header.Get("X-Content-Type-Options"); cto == "" {
		warnings = append(warnings, "missing X-Content-Type-Options header")
	}

	// Check for X-Frame-Options
	if xfo := resp.Header.Get("X-Frame-Options"); xfo == "" {
		warnings = append(warnings, "missing X-Frame-Options header")
	}

	// Check for X-XSS-Protection
	if xxp := resp.Header.Get("X-XSS-Protection"); xxp == "" {
		warnings = append(warnings, "missing X-XSS-Protection header")
	}

	// Check for Strict-Transport-Security (for HTTPS responses)
	if resp.Request.URL.Scheme == "https" {
		if sts := resp.Header.Get("Strict-Transport-Security"); sts == "" {
			warnings = append(warnings, "missing Strict-Transport-Security header for HTTPS")
		}
	}

	if len(warnings) > 0 {
		return fmt.Errorf("%w: %s", ErrSecurityHeaderValidation, strings.Join(warnings, "; "))
	}

	return nil
}

// ExecuteDownloadRequest executes the HTTP request with retry logic and returns the response.
// It closes the response body immediately for non-OK status codes to prevent resource leaks.
// Implements exponential backoff for retries, attempting downloads up to maxRetries times.
func (d *Downloader) executeDownloadRequest(req *http.Request) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)

	// Loop through download attempts with exponential backoff
	for attempt := 0; attempt <= maxRetries; attempt++ {
		resp, err = d.client.Do(req)
		if err != nil {
			if attempt < maxRetries {
				logger.Debugf("Download attempt %d failed: %v, retrying in %v", attempt+1, err, retryDelay*time.Duration(attempt+1))
				time.Sleep(retryDelay * time.Duration(attempt+1))

				continue
			}

			return nil, &NetworkError{
				StatusCode: 0,
				URL:        req.URL.String(),
				Response:   "",
				Err:        fmt.Errorf("failed to download after %d attempts: %w", maxRetries+1, err),
			}
		}

		// Validate security headers in response
		err := validateSecurityHeaders(resp)
		if err != nil {
			logger.Warnf("Security header validation failed: %v", err)
			// Continue processing but log the issue
		}

		// Check if the response status indicates success
		if resp.StatusCode == http.StatusOK {
			return resp, nil
		}

		// Close response body immediately for non-OK status codes to prevent resource leaks
		_ = resp.Body.Close()

		// Retry on server errors (5xx) or rate limiting (429 Too Many Requests)
		if resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests {
			if attempt < maxRetries {
				logger.Debugf("Download attempt %d failed with status %d, retrying in %v",
					attempt+1, resp.StatusCode, retryDelay*time.Duration(attempt+1))
				time.Sleep(retryDelay * time.Duration(attempt+1))

				continue
			}
		}

		return nil, &NetworkError{
			StatusCode: resp.StatusCode,
			URL:        req.URL.String(),
			Response:   "",
			Err:        ErrDownloadFailed,
		}
	}

	return resp, nil
}

// NewDefaultHTTPClient creates a new DefaultHTTPClient with an initialized HTTP client.
// It returns a pointer to DefaultHTTPClient.
func NewDefaultHTTPClient() *DefaultHTTPClient {
	return &DefaultHTTPClient{
		client: httpclient.NewHTTPClient(),
	}
}

// Do implements the HTTPClient interface by using the pre-initialized HTTP client.
func (d *DefaultHTTPClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}

	return resp, nil
}
