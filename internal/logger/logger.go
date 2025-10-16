// Package logger provides logging utilities for goUpdater.
// It uses zerolog for structured logging with configurable verbosity.
package logger

import (
	"io"
	"strings"

	"github.com/rs/zerolog"
)

//nolint:gochecknoglobals
var (
	logger  = zerolog.New(createConsoleWriter()).With().Timestamp().Logger()
	verbose bool
)

// createConsoleWriter creates a custom console writer with conditional formatting.
// Non-verbose: timestamp and message only (no level prefix)
// Verbose: timestamp, level prefix, and message
// Error/Warning: always show timestamp, level prefix, and message.
func createConsoleWriter() zerolog.ConsoleWriter {
	writer := zerolog.NewConsoleWriter()
	writer.FormatLevel = formatLevel

	return writer
}

// formatLevel formats the log level based on verbosity and level type.
// Error and Warning levels always show the prefix.
// Info level shows prefix only in verbose mode.
// Debug level shows prefix only in verbose mode.
//
//nolint:cyclop
func formatLevel(i any) string {
	level, ok := i.(string)
	if !ok {
		return ""
	}

	switch level {
	case "error", "fatal", "panic":
		return "Error: "
	case "warn":
		return "Warning: "
	case "info":
		if verbose {
			return "Info: "
		}

		return ""
	case "debug":
		if verbose {
			return "Debug: "
		}

		return ""
	case "trace":
		if verbose {
			return "Trace: "
		}

		return ""
	default:
		return strings.ToUpper(level + ": ")
	}
}

// Info logs an informational message.
// The args parameter is unused for compatibility but not utilized in the current implementation.
func Info(msg string, _ ...any) {
	logger.Info().Msg(msg)
}

// Error logs an error message.
// The args parameter is unused for compatibility but not utilized in the current implementation.
func Error(msg string, _ ...any) {
	logger.Error().Msg(msg)
}

// Infof logs a formatted informational message.
func Infof(format string, args ...any) {
	logger.Info().Msgf(format, args...)
}

// Errorf logs a formatted error message.
func Errorf(format string, args ...any) {
	logger.Error().Msgf(format, args...)
}

// Debug logs a debug message when verbose mode is enabled.
// The args parameter is unused for compatibility but not utilized in the current implementation.
func Debug(msg string, _ ...any) {
	logger.Debug().Msg(msg)
}

// Debugf logs a formatted debug message when verbose mode is enabled.
func Debugf(format string, args ...any) {
	logger.Debug().Msgf(format, args...)
}

// Warn logs a warning message.
// The args parameter is unused for compatibility but not utilized in the current implementation.
func Warn(msg string, _ ...any) {
	logger.Warn().Msg(msg)
}

// Warnf logs a formatted warning message.
func Warnf(format string, args ...any) {
	logger.Warn().Msgf(format, args...)
}

// SetVerbose sets the global log level to enable or disable debug logging.
// When verbose is true, debug messages are logged; otherwise, only info and above are logged.
// Also updates the global verbose flag for custom formatting.
func SetVerbose(v bool) {
	verbose = v
	if v {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// SetWriter sets the output writer for the logger.
// This is primarily used for testing to capture log output.
func SetWriter(w io.Writer) {
	consoleWriter := createConsoleWriter()
	consoleWriter.Out = w
	logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
}
