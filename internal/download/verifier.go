// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"bufio"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

// verifyChecksum computes the SHA256 checksum of the file using streaming I/O and compares it to the expected value.
// This approach avoids loading the entire file into memory, making it suitable for large archives.
// It uses buffered reading for better performance with large files.
func (d *Downloader) verifyChecksum(filePath, expectedSha256 string) error {
	logger.Debugf("Verifying checksum for file: %s", filePath)

	file, err := d.fs.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	defer func() { _ = file.Close() }()

	hasher := sha256.New()

	logger.Debug("Computing SHA256 hash using streaming I/O")

	// Use buffered reader for better performance with large files
	bufferedReader := bufio.NewReaderSize(file, 64*1024) //nolint:mnd // 64KB buffer

	_, err = io.Copy(hasher, bufferedReader)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	actualSha256 := hex.EncodeToString(hasher.Sum(nil))
	logger.Debugf("Computed hash: %s", actualSha256)

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(actualSha256), []byte(expectedSha256)) != 1 {
		return &ChecksumError{
			FilePath:       filePath,
			ExpectedSha256: expectedSha256,
			ActualSha256:   actualSha256,
			Err:            ErrChecksumMismatch,
		}
	}

	logger.Debug("Checksum verification passed")

	return nil
}
