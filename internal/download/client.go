// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/nicholas-fedor/goUpdater/internal/logger"
	"github.com/nicholas-fedor/goUpdater/internal/version"
)

// validateDownloadURL validates the URL for security and correctness before creating a download request.
// It ensures the URL is well-formed, uses HTTPS, and doesn't contain dangerous schemes or paths.
//
//nolint:cyclop // Complex validation logic requires multiple security checks
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
	host := strings.ToLower(parsedURL.Host)
	if strings.Contains(host, "localhost") || strings.Contains(host, "127.0.0.1") ||
		strings.Contains(host, "0.0.0.0") || strings.HasPrefix(host, "10.") ||
		strings.HasPrefix(host, "172.") || strings.HasPrefix(host, "192.168.") {
		return ErrInvalidURLHost
	}

	// Basic path validation - prevent directory traversal
	if strings.Contains(parsedURL.Path, "..") {
		return ErrDirectoryTraversal
	}

	return nil
}

// CreateDownloadRequest creates an HTTP GET request for the given URL with context and proper headers.
// It sets appropriate headers for modern HTTP clients to ensure compatibility and proper content negotiation.
func (d *Downloader) createDownloadRequest(url string) (*http.Request, error) {
	logger.Debugf("Creating HTTP request for: %s", url)

	// Validate URL for security
	err := validateDownloadURL(url)
	if err != nil {
		return nil, fmt.Errorf("URL validation failed: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set proper headers for modern HTTP clients
	req.Header.Set("User-Agent", "goUpdater/"+version.GetClientVersion())
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")

	// Security headers
	req.Header.Set("X-Content-Type-Options", "nosniff")
	req.Header.Set("X-Frame-Options", "DENY")
	req.Header.Set("X-XSS-Protection", "1; mode=block")

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
// It ensures the response body is closed on error and implements exponential backoff for retries.
// The function attempts the download up to maxRetries times, handling network errors and server-side issues.
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

		// Close response body for non-OK status codes to prevent resource leaks
		defer func() { _ = resp.Body.Close() }()

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
