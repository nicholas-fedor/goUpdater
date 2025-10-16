// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestInfo(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	SetWriter(&buf)

	Info("test message")

	output := strings.TrimSpace(buf.String())

	// Check if output contains the message and timestamp (no level prefix in non-verbose mode)
	if strings.Contains(output, "Info:") || !strings.Contains(output, "test message") {
		t.Errorf("Expected output to contain 'test message' without 'Info:' prefix in non-verbose mode, got %s", output)
	}
}

func TestError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	SetWriter(&buf)

	Error("test error message")

	output := strings.TrimSpace(buf.String())

	// Check if output contains the message and Error prefix (always shown)
	if !strings.Contains(output, "Error:") || !strings.Contains(output, "test error message") {
		t.Errorf("Expected output to contain 'Error:' and 'test error message', got %s", output)
	}
}

func TestInfof(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	SetWriter(&buf)

	Infof("test %s message", "formatted")

	output := strings.TrimSpace(buf.String())

	// Check if output contains the message without Info prefix in non-verbose mode
	if strings.Contains(output, "Info:") || !strings.Contains(output, "test formatted message") {
		t.Errorf("Expected output to contain 'test formatted message' without 'Info:' prefix, got %s", output)
	}
}

func TestErrorf(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	SetWriter(&buf)

	Errorf("test %s error", "formatted")

	output := strings.TrimSpace(buf.String())

	// Check if output contains the message and Error prefix (always shown)
	if !strings.Contains(output, "Error:") || !strings.Contains(output, "test formatted error") {
		t.Errorf("Expected output to contain 'Error:' and 'test formatted error', got %s", output)
	}
}

func TestDebug(t *testing.T) {
	t.Parallel()
	// Set verbose to true to enable debug logs
	SetVerbose(true)

	var buf bytes.Buffer
	SetWriter(&buf)

	Debug("test debug message")

	output := strings.TrimSpace(buf.String())

	// Check if output contains the message and Debug prefix in verbose mode
	if !strings.Contains(output, "Debug:") || !strings.Contains(output, "test debug message") {
		t.Errorf("Expected output to contain 'Debug:' and 'test debug message' in verbose mode, got %s", output)
	}

	// Reset to default
	SetVerbose(false)
}

func TestDebugf(t *testing.T) {
	t.Parallel()
	// Set verbose to true to enable debug logs
	SetVerbose(true)

	var buf bytes.Buffer
	SetWriter(&buf)

	Debugf("test %s debug", "formatted")

	output := strings.TrimSpace(buf.String())

	// Check if output contains the message and Debug prefix in verbose mode
	if !strings.Contains(output, "Debug:") || !strings.Contains(output, "test formatted debug") {
		t.Errorf("Expected output to contain 'Debug:' and 'test formatted debug' in verbose mode, got %s", output)
	}

	// Reset to default
	SetVerbose(false)
}

func TestSetVerbose(t *testing.T) {
	t.Parallel()
	// Test setting verbose to true
	SetVerbose(true)

	var buf bytes.Buffer
	SetWriter(&buf)

	Debug("verbose enabled")

	output := strings.TrimSpace(buf.String())

	// Check if output contains the message and Debug prefix in verbose mode
	if !strings.Contains(output, "Debug:") || !strings.Contains(output, "verbose enabled") {
		t.Errorf("Expected debug output with 'Debug:' prefix when verbose is enabled, got %s", output)
	}

	// Test setting verbose to false
	SetVerbose(false)

	var buf2 bytes.Buffer
	SetWriter(&buf2)

	Debug("verbose disabled")

	output2 := strings.TrimSpace(buf2.String())

	// When verbose is false, debug logs should not appear
	if strings.Contains(output2, "verbose disabled") {
		t.Errorf("Expected no debug output when verbose is disabled, got %s", output2)
	}
}

func TestWarn(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	SetWriter(&buf)

	Warn("test warning message")

	output := strings.TrimSpace(buf.String())

	// Check if output contains the message and Warning prefix (always shown)
	if !strings.Contains(output, "Warning:") || !strings.Contains(output, "test warning message") {
		t.Errorf("Expected output to contain 'Warning:' and 'test warning message', got %s", output)
	}
}

func TestWarnf(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	SetWriter(&buf)

	Warnf("test %s warning", "formatted")

	output := strings.TrimSpace(buf.String())

	// Check if output contains the message and Warning prefix (always shown)
	if !strings.Contains(output, "Warning:") || !strings.Contains(output, "test formatted warning") {
		t.Errorf("Expected output to contain 'Warning:' and 'test formatted warning', got %s", output)
	}
}

func TestInfoVerbose(t *testing.T) {
	t.Parallel()
	// Set verbose to true to enable info level prefix
	SetVerbose(true)

	var buf bytes.Buffer
	SetWriter(&buf)

	Info("test info message")

	output := strings.TrimSpace(buf.String())

	// Check if output contains the message and Info prefix in verbose mode
	if !strings.Contains(output, "Info:") || !strings.Contains(output, "test info message") {
		t.Errorf("Expected output to contain 'Info:' and 'test info message' in verbose mode, got %s", output)
	}

	// Reset to default
	SetVerbose(false)
}
