// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package archive

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultProcessor_NewGzipReader(t *testing.T) {
	t.Parallel()

	// Test successful creation of gzip reader with valid gzip data
	t.Run("successfulGzipReaderCreation", func(t *testing.T) {
		t.Parallel()
		// Create a buffer with valid gzip data
		var buf bytes.Buffer

		gzipWriter := gzip.NewWriter(&buf)
		_, err := gzipWriter.Write([]byte("test data"))
		require.NoError(t, err)
		require.NoError(t, gzipWriter.Close())

		processor := &DefaultProcessor{}
		reader, err := processor.NewGzipReader(&buf)

		// Assert no error occurred
		require.NoError(t, err)
		// Assert reader is not nil
		assert.NotNil(t, reader)
		// Assert reader can read the data
		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, "test data", string(data))
		// Close the reader
		require.NoError(t, reader.Close())
	})

	// Test error case with invalid gzip data
	t.Run("errorWithInvalidGzipData", func(t *testing.T) {
		t.Parallel()
		// Create a reader with invalid gzip data
		invalidReader := strings.NewReader("invalid gzip data")

		processor := &DefaultProcessor{}
		reader, err := processor.NewGzipReader(invalidReader)

		// Assert error occurred
		require.Error(t, err)
		// Assert reader is nil
		assert.Nil(t, reader)
		// Assert error message contains expected text
		assert.Contains(t, err.Error(), "failed to create gzip reader")
	})
}

func TestDefaultProcessor_NewTarReader(t *testing.T) {
	t.Parallel()

	// Test successful creation of a tar reader with valid tar data
	t.Run("successfulTarReaderCreation", func(t *testing.T) {
		t.Parallel()
		// Create a buffer with valid tar data
		var buf bytes.Buffer

		tarWriter := tar.NewWriter(&buf)
		header := &tar.Header{
			Name: "test.txt",
			Mode: 0600,
			Size: int64(len("test data")),
		}
		require.NoError(t, tarWriter.WriteHeader(header))
		_, err := tarWriter.Write([]byte("test data"))
		require.NoError(t, err)
		require.NoError(t, tarWriter.Close())

		processor := &DefaultProcessor{}
		reader := processor.NewTarReader(&buf)

		// Assert reader is not nil
		assert.NotNil(t, reader)
		// Assert reader implements TarReader interface
		assert.Implements(t, (*TarReader)(nil), reader)
		// Assert reader can read the header
		hdr, err := reader.Next()
		require.NoError(t, err)
		assert.Equal(t, "test.txt", hdr.Name)
		assert.Equal(t, int64(len("test data")), hdr.Size)
		// Assert reader can read the data
		data, err := io.ReadAll(reader)
		require.NoError(t, err)
		assert.Equal(t, "test data", string(data))
		// Assert no more entries
		hdr, err = reader.Next()
		assert.Equal(t, io.EOF, err)
		assert.Nil(t, hdr)
	})

	// Test creation of tar reader with empty reader
	t.Run("tarReaderWithEmptyReader", func(t *testing.T) {
		t.Parallel()
		// Create an empty reader
		emptyReader := strings.NewReader("")

		processor := &DefaultProcessor{}
		reader := processor.NewTarReader(emptyReader)

		// Assert reader is not nil
		assert.NotNil(t, reader)
		// Assert reader implements TarReader interface
		assert.Implements(t, (*TarReader)(nil), reader)
		// Assert reader returns EOF immediately
		hdr, err := reader.Next()
		assert.Equal(t, io.EOF, err)
		assert.Nil(t, hdr)
	})

	// Test creation of tar reader with invalid tar data
	t.Run("tarReaderWithInvalidData", func(t *testing.T) {
		t.Parallel()
		// Create a reader with invalid tar data
		invalidReader := strings.NewReader("invalid tar data")

		processor := &DefaultProcessor{}
		reader := processor.NewTarReader(invalidReader)

		// Assert reader is not nil
		assert.NotNil(t, reader)
		// Assert reader implements TarReader interface
		assert.Implements(t, (*TarReader)(nil), reader)
		// Assert reader returns error on Next()
		hdr, err := reader.Next()
		require.Error(t, err)
		assert.Nil(t, hdr)
	})
}
