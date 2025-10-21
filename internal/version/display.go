// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package version

import (
	"fmt"
	"io"

	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
)

// NewDisplayManager creates a new version Display Manager with the provided dependencies.
// It initializes a Display Manager struct with the given reader, parser, encoder, and writer.
// The reader is used for reading debug build information, parser for time parsing,
// encoder for JSON encoding, and writer for error output.
func NewDisplayManager(
	reader filesystem.DebugInfoReader,
	parser filesystem.TimeParser,
	encoder filesystem.JSONEncoder,
	writer filesystem.ErrorWriter,
) *DisplayManager {
	return &DisplayManager{
		reader:  reader,
		parser:  parser,
		encoder: encoder,
		writer:  writer,
	}
}

// DisplayDefault displays version information in the default format.
// It outputs "goUpdater <version>" only.
func (dm *DisplayManager) DisplayDefault(writer io.Writer, info Info) {
	if info.Version != "" {
		_, _ = fmt.Fprintf(writer, "goUpdater %s\n", info.Version)
	} else {
		_, _ = fmt.Fprintf(writer, "goUpdater\n")
	}
}

// DisplayShort displays version information in short format.
// It outputs only the version number.
func (dm *DisplayManager) DisplayShort(writer io.Writer, info Info) {
	_, _ = fmt.Fprintf(writer, "%s\n", info.Version)
}

// DisplayVerbose displays version information in verbose format.
// It shows all available information in a tree-like structure.
func (dm *DisplayManager) DisplayVerbose(writer io.Writer, info Info) {
	_, _ = fmt.Fprintf(writer, "goUpdater\n")

	if info.Version != "" {
		_, _ = fmt.Fprintf(writer, "├─ Version: %s\n", info.Version)
	}

	if info.Commit != "" {
		_, _ = fmt.Fprintf(writer, "├─ Commit: %s\n", info.Commit)
	}

	if info.Date != "" {
		_, _ = fmt.Fprintf(writer, "├─ Date: %s\n", info.Date)
	}

	if info.GoVersion != "" {
		_, _ = fmt.Fprintf(writer, "├─ Go Version: %s\n", info.GoVersion)
	}

	if info.Platform != "" {
		_, _ = fmt.Fprintf(writer, "└─ Platform: %s\n", info.Platform)
	}
}

// DisplayJSON displays version information in JSON format.
// It encodes the version information as JSON to the writer.
func (dm *DisplayManager) DisplayJSON(writer io.Writer, info Info) error {
	encoder := dm.encoder.NewEncoder(writer)
	if encoder == nil {
		return ErrFailedToCreateEncoder
	}

	err := encoder.Encode(info)
	if err != nil {
		return fmt.Errorf("failed to encode version info to JSON: %w", err)
	}

	return nil
}

// DisplayWithFormat displays version information in the specified format.
// It supports "default", "short", "verbose", and "json" formats.
func DisplayWithFormat(writer io.Writer, format string) {
	displayManager := NewDisplayManager(
		&filesystem.OSDebugInfoReader{},
		&filesystem.OSTimeParser{},
		&filesystem.OSJSONEncoder{},
		&filesystem.OSErrorWriter{},
	)
	info := GetVersionInfo(&filesystem.OSDebugInfoReader{})

	switch format {
	case "short":
		displayManager.DisplayShort(writer, info)
	case "verbose":
		displayManager.DisplayVerbose(writer, info)
	case "json":
		_ = displayManager.DisplayJSON(writer, info)
	default:
		displayManager.DisplayDefault(writer, info)
	}
}
