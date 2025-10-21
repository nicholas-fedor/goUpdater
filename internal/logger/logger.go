// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package logger provides logging utilities for goUpdater.
// It uses zerolog for structured logging with configurable verbosity.
package logger

import (
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

var (
	logger  = createLogger()
	verbose bool
	mutex   sync.RWMutex
)

// createConsoleWriter creates a custom console writer with conditional formatting.
// Non-verbose: message only (no timestamp, no level prefix)
// Verbose: timestamp, level prefix, and message
// Error/Warning: always show timestamp, level prefix, and message.
func createConsoleWriter() zerolog.ConsoleWriter {
	writer := zerolog.NewConsoleWriter()
	writer.TimeFormat = "2006-01-02T15:04:05-07:00"
	writer.FormatTimestamp = func(i interface{}) string {
		return fmt.Sprintf("\x1b[90m%s\x1b[0m", i)
	}

	writer.FormatLevel = formatLevel
	if verbose {
		writer.PartsOrder = []string{"time", "level", "message"}
	} else {
		writer.PartsOrder = []string{"level", "message"}
	}

	return writer
}

// createLogger creates a logger with conditional timestamp inclusion.
// Timestamps are included when verbose mode is enabled.
func createLogger() zerolog.Logger {
	consoleWriter := createConsoleWriter()
	if verbose {
		return zerolog.New(consoleWriter).With().Timestamp().Logger()
	}

	return zerolog.New(consoleWriter)
}

// formatLevel formats the log level based on verbosity and level type.
// Error and Warning levels always show the prefix.
// Info level shows prefix only in verbose mode.
// Debug level shows prefix only in verbose mode.
// In non-verbose mode, info messages have no prefix for clean output.
func formatLevel(i any) string {
	level, ok := i.(string)
	if !ok {
		return ""
	}

	switch level {
	case "error", "fatal", "panic":
		return "Error:"
	case "warn":
		return "Warning:"
	case "info":
		return ""
	case "debug":
		if verbose {
			return "Debug:"
		}

		return ""
	default:
		return strings.ToUpper(level + ":")
	}
}

// Info logs an informational message.
// The args parameter is unused for compatibility but not utilized in the current implementation.
func Info(msg string, _ ...any) {
	mutex.RLock()

	l := logger

	mutex.RUnlock()
	l.Info().Msg(msg)
}

// Error logs an error message.
// The args parameter is unused for compatibility but not utilized in the current implementation.
func Error(msg string, _ ...any) {
	mutex.RLock()

	l := logger

	mutex.RUnlock()
	l.Error().Msg(msg)
}

// Infof logs a formatted informational message.
func Infof(format string, args ...any) {
	mutex.RLock()

	l := logger

	mutex.RUnlock()
	l.Info().Msgf(format, args...)
}

// Errorf logs a formatted error message.
func Errorf(format string, args ...any) {
	mutex.RLock()

	l := logger

	mutex.RUnlock()
	l.Error().Msgf(format, args...)
}

// Debug logs a debug message when verbose mode is enabled.
// The args parameter is unused for compatibility but not utilized in the current implementation.
func Debug(msg string, _ ...any) {
	mutex.RLock()

	l := logger

	mutex.RUnlock()
	l.Debug().Msg(msg)
}

// Debugf logs a formatted debug message when verbose mode is enabled.
func Debugf(format string, args ...any) {
	mutex.RLock()

	l := logger

	mutex.RUnlock()
	l.Debug().Msgf(format, args...)
}

// Warn logs a warning message.
// The args parameter is unused for compatibility but not utilized in the current implementation.
func Warn(msg string, _ ...any) {
	mutex.RLock()

	l := logger

	mutex.RUnlock()
	l.Warn().Msg(msg)
}

// Warnf logs a formatted warning message.
func Warnf(format string, args ...any) {
	mutex.RLock()

	l := logger

	mutex.RUnlock()
	l.Warn().Msgf(format, args...)
}

// SetVerbose sets the global log level to enable or disable debug logging.
// When verbose is true, debug messages are logged; otherwise, only info and above are logged.
// Also updates the global verbose flag for custom formatting and recreates logger.
func SetVerbose(v bool) {
	mutex.Lock()

	verbose = v
	if v {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	logger = createLogger()

	mutex.Unlock()
}

// SetWriter sets the output writer for the logger.
// This is primarily used for testing to capture log output.
func SetWriter(w io.Writer) {
	mutex.Lock()

	consoleWriter := createConsoleWriter()
	consoleWriter.Out = w

	logger = zerolog.New(consoleWriter)
	if verbose {
		logger = logger.With().Timestamp().Logger()
	}

	mutex.Unlock()
}
