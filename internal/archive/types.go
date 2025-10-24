// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package archive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"

	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
)

// Processor handles archive extraction operations.
type Processor interface {
	NewGzipReader(r io.Reader) (io.ReadCloser, error)
	NewTarReader(r io.Reader) TarReader
}

// DefaultProcessor provides default implementation of Processor.
type DefaultProcessor struct{}

// TarReader provides tar archive reading functionality.
type TarReader interface {
	Next() (*tar.Header, error)
	Read(b []byte) (int, error)
}

// TarHeader represents a tar header.
type TarHeader = tar.Header

// Extractor handles archive extraction with dependency injection.
type Extractor struct {
	fs        filesystem.FileSystem
	processor Processor
}

// NewGzipReader creates a new gzip reader.
func (d *DefaultProcessor) NewGzipReader(r io.Reader) (io.ReadCloser, error) {
	reader, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}

	return reader, nil
}

// NewTarReader creates a new tar reader.
func (d *DefaultProcessor) NewTarReader(r io.Reader) TarReader {
	return tar.NewReader(r)
}
