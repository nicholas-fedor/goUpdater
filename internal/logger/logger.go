// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package logger provides logging utilities for goUpdater.
// It uses zerolog for structured logging with configurable verbosity.
package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

// Logger encapsulates logging state and provides logging methods.
// It eliminates global state by maintaining its own zerolog.Logger instance.
type Logger struct {
	logger  zerolog.Logger
	verbose bool
	writer  io.Writer
	mutex   sync.RWMutex
}

// NewLogger creates a new Logger instance with the specified verbosity and output writer.
func NewLogger(verbose bool, writer io.Writer) *Logger {
	//nolint:exhaustruct // fields initialized via zero values and updateLogger
	logger := &Logger{
		verbose: verbose,
		writer:  writer,
	}
	logger.updateLogger()

	return logger
}

// Info logs an informational message using the Logger instance.
func (l *Logger) Info(msg string, _ ...any) {
	l.mutex.RLock()
	l.logger.Info().Msg(msg)
	l.mutex.RUnlock()
}

// Error logs an error message using the Logger instance.
func (l *Logger) Error(msg string, _ ...any) {
	l.mutex.RLock()
	l.logger.Error().Msg(msg)
	l.mutex.RUnlock()
}

// Infof logs a formatted informational message using the Logger instance.
func (l *Logger) Infof(format string, args ...any) {
	l.mutex.RLock()
	l.logger.Info().Msgf(format, args...)
	l.mutex.RUnlock()
}

// Errorf logs a formatted error message using the Logger instance.
func (l *Logger) Errorf(format string, args ...any) {
	l.mutex.RLock()
	l.logger.Error().Msgf(format, args...)
	l.mutex.RUnlock()
}

// Debug logs a debug message when verbose mode is enabled using the Logger instance.
func (l *Logger) Debug(msg string, _ ...any) {
	l.mutex.RLock()
	l.logger.Debug().Msg(msg)
	l.mutex.RUnlock()
}

// Debugf logs a formatted debug message when verbose mode is enabled using the Logger instance.
func (l *Logger) Debugf(format string, args ...any) {
	l.mutex.RLock()
	l.logger.Debug().Msgf(format, args...)
	l.mutex.RUnlock()
}

// Warn logs a warning message using the Logger instance.
func (l *Logger) Warn(msg string, _ ...any) {
	l.mutex.RLock()
	l.logger.Warn().Msg(msg)
	l.mutex.RUnlock()
}

// Warnf logs a formatted warning message using the Logger instance.
func (l *Logger) Warnf(format string, args ...any) {
	l.mutex.RLock()
	l.logger.Warn().Msgf(format, args...)
	l.mutex.RUnlock()
}

// SetVerbose sets the verbosity for the Logger instance.
// When verbose is true, debug messages are logged; otherwise, only info and above are logged.
func (l *Logger) SetVerbose(v bool) {
	l.mutex.Lock()
	l.verbose = v
	l.updateLogger()
	l.mutex.Unlock()
}

// SetWriter sets the output writer for the Logger instance.
func (l *Logger) SetWriter(w io.Writer) {
	l.mutex.Lock()
	l.writer = w
	l.updateLogger()
	l.mutex.Unlock()
}

// updateLogger recreates the internal zerolog.Logger based on current verbose and writer settings.
func (l *Logger) updateLogger() {
	consoleWriter := l.createConsoleWriter()

	level := zerolog.InfoLevel
	if l.verbose {
		level = zerolog.DebugLevel
	}

	l.logger = zerolog.New(consoleWriter).Level(level)
	if l.verbose {
		l.logger = l.logger.With().Timestamp().Logger()
	}
}

// createConsoleWriter creates a custom console writer with conditional formatting.
// Non-verbose: message only (no timestamp, no level prefix)
// Verbose: timestamp, level prefix, and message
// Error/Warning: always show timestamp, level prefix, and message.
func (l *Logger) createConsoleWriter() zerolog.ConsoleWriter {
	writer := zerolog.NewConsoleWriter()
	writer.Out = l.writer
	writer.TimeFormat = "2006-01-02T15:04:05-07:00"
	writer.FormatTimestamp = func(i interface{}) string {
		return fmt.Sprintf("\x1b[90m%s\x1b[0m", i)
	}

	writer.FormatLevel = l.formatLevel
	if l.verbose {
		writer.PartsOrder = []string{"time", "level", "message"}
	} else {
		writer.PartsOrder = []string{"level", "message"}
	}

	return writer
}

// formatLevel formats the log level based on verbosity and level type.
// Error and Warning levels always show the prefix.
// Info level shows prefix only in verbose mode.
// Debug level shows prefix only in verbose mode.
// In non-verbose mode, info messages have no prefix for clean output.
func (l *Logger) formatLevel(i any) string {
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
		if l.verbose {
			return "Debug:"
		}

		return ""
	default:
		return strings.ToUpper(level + ":")
	}
}

var globalLogger = NewLogger(false, os.Stderr)

// Info logs an informational message using the global logger.
func Info(msg string, args ...any) {
	globalLogger.Info(msg, args...)
}

// Error logs an error message using the global logger.
func Error(msg string, args ...any) {
	globalLogger.Error(msg, args...)
}

// Infof logs a formatted informational message using the global logger.
func Infof(format string, args ...any) {
	globalLogger.Infof(format, args...)
}

// Errorf logs a formatted error message using the global logger.
func Errorf(format string, args ...any) {
	globalLogger.Errorf(format, args...)
}

// Debug logs a debug message when verbose mode is enabled using the global logger.
func Debug(msg string, args ...any) {
	globalLogger.Debug(msg, args...)
}

// Debugf logs a formatted debug message when verbose mode is enabled using the global logger.
func Debugf(format string, args ...any) {
	globalLogger.Debugf(format, args...)
}

// Warn logs a warning message using the global logger.
func Warn(msg string, args ...any) {
	globalLogger.Warn(msg, args...)
}

// Warnf logs a formatted warning message using the global logger.
func Warnf(format string, args ...any) {
	globalLogger.Warnf(format, args...)
}

// SetVerbose sets the verbosity for the global logger.
func SetVerbose(v bool) {
	globalLogger.SetVerbose(v)
}

// SetWriter sets the output writer for the global logger.
func SetWriter(w io.Writer) {
	globalLogger.SetWriter(w)
}
