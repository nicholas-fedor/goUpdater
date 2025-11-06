// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package logger

import (
	"bytes"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// stripANSI removes ANSI escape codes from a string.
func stripANSI(str string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[mG]`)

	return ansiRegex.ReplaceAllString(str, "")
}

// stripTimestamp removes timestamp from log lines.
func stripTimestamp(str string) string {
	timestampRegex := regexp.MustCompile(`\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[+-]\d{2}:\d{2}\s+`)

	return timestampRegex.ReplaceAllString(str, "")
}

func TestInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		message  string
		verbose  bool
		expected string
	}{
		{
			name:     "info message non-verbose",
			message:  "test info",
			verbose:  false,
			expected: "test info\n",
		},
		{
			name:     "info message verbose",
			message:  "test info",
			verbose:  true,
			expected: "test info",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			logger := NewLogger(testCase.verbose, buf)

			logger.Info(testCase.message)

			output := stripTimestamp(stripANSI(buf.String()))
			t.Logf("Test %s: raw=%q, ansi_stripped=%q, timestamp_stripped=%q, expected=%q",
				testCase.name, buf.String(), stripANSI(buf.String()), output, testCase.expected)
			assert.Contains(t, output, testCase.expected)
		})
	}
}

func TestError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		message  string
		verbose  bool
		expected string
	}{
		{
			name:     "error message non-verbose",
			message:  "test error",
			verbose:  false,
			expected: "Error: test error",
		},
		{
			name:     "error message verbose",
			message:  "test error",
			verbose:  true,
			expected: "Error: test error",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			logger := NewLogger(testCase.verbose, buf)

			logger.Error(testCase.message)

			output := stripTimestamp(stripANSI(buf.String()))
			t.Logf("Test %s: raw=%q, ansi_stripped=%q, timestamp_stripped=%q, expected=%q",
				testCase.name, buf.String(), stripANSI(buf.String()), output, testCase.expected)
			assert.Contains(t, output, testCase.expected)
		})
	}
}

func TestWarn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		message  string
		verbose  bool
		expected string
	}{
		{
			name:     "warn message non-verbose",
			message:  "test warn",
			verbose:  false,
			expected: "Warning: test warn\n",
		},
		{
			name:     "warn message verbose",
			message:  "test warn",
			verbose:  true,
			expected: "Warning: test warn\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			logger := NewLogger(testCase.verbose, buf)

			logger.Warn(testCase.message)

			output := stripTimestamp(stripANSI(buf.String()))
			assert.Contains(t, output, testCase.expected)
		})
	}
}

func TestDebug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		message  string
		verbose  bool
		expected string
	}{
		{
			name:     "debug message non-verbose",
			message:  "test debug",
			verbose:  false,
			expected: "", // Debug should not output in non-verbose
		},
		{
			name:     "debug message verbose",
			message:  "test debug",
			verbose:  true,
			expected: "Debug: test debug\n",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			logger := NewLogger(testCase.verbose, buf)

			logger.Debug(testCase.message)

			output := stripTimestamp(stripANSI(buf.String()))
			if testCase.expected == "" {
				assert.Empty(t, output)
			} else {
				assert.Contains(t, output, testCase.expected)
			}
		})
	}
}

func TestLoggingFunctions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		verbose  bool
		callType string
		expected string
	}{
		{
			name:     "Infof",
			verbose:  false,
			callType: "Infof",
			expected: "test infof 123\n",
		},
		{
			name:     "Errorf",
			verbose:  false,
			callType: "Errorf",
			expected: "Error: test errorf 456\n",
		},
		{
			name:     "Warnf",
			verbose:  false,
			callType: "Warnf",
			expected: "Warning: test warnf 789\n",
		},
		{
			name:     "Debugf",
			verbose:  true,
			callType: "Debugf",
			expected: "Debug: test debugf 101\n",
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			logger := NewLogger(testCase.verbose, buf)

			switch testCase.callType {
			case "Infof":
				logger.Infof("test %s %d", "infof", 123)
			case "Errorf":
				logger.Errorf("test %s %d", "errorf", 456)
			case "Warnf":
				logger.Warnf("test %s %d", "warnf", 789)
			case "Debugf":
				logger.Debugf("test %s %d", "debugf", 101)
			}

			output := stripTimestamp(stripANSI(buf.String()))
			assert.Contains(t, output, testCase.expected)
		})
	}
}

func TestSetVerbose(t *testing.T) {
	t.Parallel()

	// Test setting verbose to true
	buf := &bytes.Buffer{}
	logger := NewLogger(false, buf)
	logger.SetVerbose(true)

	logger.Debug("debug message")

	output := stripTimestamp(stripANSI(buf.String()))
	assert.Contains(t, output, "Debug: debug message\n")

	// Test setting verbose to false
	buf2 := &bytes.Buffer{}
	logger2 := NewLogger(true, buf2)
	logger2.SetVerbose(false)

	logger2.Debug("debug message")

	output2 := stripTimestamp(stripANSI(buf2.String()))
	assert.Empty(t, output2)
}

func TestSetWriter(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	logger := NewLogger(false, os.Stderr)
	logger.SetWriter(buf)

	logger.Info("test writer")

	output := stripTimestamp(stripANSI(buf.String()))
	assert.Contains(t, output, "test writer\n")
}

func TestConcurrentLogging(t *testing.T) {
	t.Parallel()

	// Test concurrent logging with multiple logger instances
	done := make(chan bool, 10)

	for goroutineID := range 10 {
		go func(id int) {
			buf := &bytes.Buffer{}
			logger := NewLogger(true, buf)
			logger.Infof("Concurrent log message %d", id)

			done <- true
		}(goroutineID)
	}

	// Wait for all goroutines to complete
	for range 10 {
		<-done
	}
}

func TestEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		action  func(logger *Logger)
		verbose bool
		check   func(t *testing.T, output string)
	}{
		{
			name:    "empty message info",
			action:  func(logger *Logger) { logger.Info("") },
			verbose: false,
			check: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "\n")
			},
		},
		{
			name:    "special characters",
			action:  func(logger *Logger) { logger.Info("test\n\t\r") },
			verbose: false,
			check: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, "test\n\t\r\n")
			},
		},
		{
			name:    "long message",
			action:  func(logger *Logger) { logger.Info(strings.Repeat("a", 1000)) },
			verbose: false,
			check: func(t *testing.T, output string) {
				t.Helper()
				assert.Contains(t, output, strings.Repeat("a", 1000)+"\n")
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			logger := NewLogger(testCase.verbose, buf)

			testCase.action(logger)

			output := stripTimestamp(stripANSI(buf.String()))
			testCase.check(t, output)
		})
	}
}

func TestLogLevelFiltering(t *testing.T) {
	t.Parallel()

	// Non-verbose: only info, warn, error should appear
	buf := &bytes.Buffer{}
	logger := NewLogger(false, buf)

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	output := stripTimestamp(stripANSI(buf.String()))
	assert.NotContains(t, output, "debug")
	assert.Contains(t, output, "info\n")
	assert.Contains(t, output, "Warning: warn\n")
	assert.Contains(t, output, "Error: error\n")

	// Verbose: all levels should appear
	buf2 := &bytes.Buffer{}
	logger2 := NewLogger(true, buf2)

	logger2.Debug("debug")
	logger2.Info("info")
	logger2.Warn("warn")
	logger2.Error("error")

	output2 := stripTimestamp(stripANSI(buf2.String()))
	assert.Contains(t, output2, "Debug: debug\n")
	assert.Contains(t, output2, "info\n")
	assert.Contains(t, output2, "Warning: warn\n")
	assert.Contains(t, output2, "Error: error\n")
}
