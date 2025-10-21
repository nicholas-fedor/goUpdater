// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/nicholas-fedor/goUpdater/internal/logger"
	"github.com/schollz/progressbar/v3"
)

// createDestinationFile creates the destination file for writing the download.
// It uses the injected filesystem interface to create the file at the specified path.
func (d *Downloader) createDestinationFile(destPath string) (*os.File, error) {
	logger.Debugf("Creating destination file: %s", destPath)

	out, err := d.fs.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	return out, nil
}

// downloadWithoutProgress copies data from the response body to the file without progress tracking.
// This method is used when the server doesn't provide content length information.
func (d *Downloader) downloadWithoutProgress(resp *http.Response, out *os.File) error {
	logger.Debug("Content length unknown, falling back to simple copy")

	_, err := io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	logger.Debug("File download completed")

	return nil
}

// downloadWithProgress sets up a progress bar and copies data with progress tracking using buffered I/O.
// It creates a progress bar that shows download speed, ETA, and completion percentage.
// The method uses buffered I/O for better memory efficiency during large downloads.
func (d *Downloader) downloadWithProgress(resp *http.Response, out *os.File, contentLength int64) error {
	logger.Debugf("Content length: %d bytes", contentLength)

	// Create progress bar with description
	bar := progressbar.NewOptions64(contentLength,
		progressbar.OptionSetDescription("Downloading Go archive"),
		progressbar.OptionSetWriter(os.Stderr), // Use stderr to avoid mixing with logs
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowCount(),
		progressbar.OptionThrottle(throttleDuration*time.Millisecond),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
	)

	logger.Debug("Starting download with progress tracking")

	// Create a progress reader that updates the bar
	progressReader := progressbar.NewReader(resp.Body, bar)

	// Use buffered I/O for better memory efficiency
	bufferedWriter := bufio.NewWriterSize(out, 64*1024) //nolint:mnd // 64KB buffer

	// Copy data with progress tracking and buffering
	_, err := io.Copy(bufferedWriter, &progressReader)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	// Flush the buffer to ensure all data is written
	err = bufferedWriter.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}

	logger.Debug("File download completed with progress tracking")

	return nil
}

// downloadFile downloads a file from the given URL to the specified path with progress tracking.
// It displays download speed, ETA, and completion percentage using a progress bar.
// The method chooses between progress tracking and simple copy based on content length availability.
func (d *Downloader) downloadFile(url, destPath string) error {
	req, err := d.createDownloadRequest(url)
	if err != nil {
		return err
	}

	resp, err := d.executeDownloadRequest(req)
	if err != nil {
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	out, err := d.createDestinationFile(destPath)
	if err != nil {
		return err
	}

	defer func() { _ = out.Close() }()

	// Get content length for progress bar
	contentLength := resp.ContentLength
	if contentLength <= 0 {
		return d.downloadWithoutProgress(resp, out)
	}

	return d.downloadWithProgress(resp, out, contentLength)
}
