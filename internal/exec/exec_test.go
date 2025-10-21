// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package exec

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type contextKey string

const testKey contextKey = "testKey"

func TestOSCommandExecutor_LookPath(t *testing.T) {
	t.Parallel()

	executor := OSCommandExecutor{}

	tests := []struct {
		name        string
		file        string
		expectError bool
	}{
		{
			name:        "non-existent executable",
			file:        "nonexistentcommand12345",
			expectError: true,
		},
		{
			name:        "empty filename",
			file:        "",
			expectError: true,
		},
		{
			name:        "path with spaces",
			file:        "command with spaces",
			expectError: true,
		},
		{
			name:        "relative path non-existent",
			file:        "./nonexistent",
			expectError: true,
		},
		{
			name:        "absolute path non-existent",
			file:        "/nonexistent/path/command",
			expectError: true,
		},
		{
			name:        "executable with special characters",
			file:        "command;rm -rf /",
			expectError: true,
		},
		{
			name:        "executable with null byte",
			file:        "command\x00evil",
			expectError: true,
		},
		{
			name:        "very long executable name",
			file:        string(make([]byte, 4097)), // PATH_MAX + 1
			expectError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			path, err := executor.LookPath(testCase.file)

			if testCase.expectError {
				require.Error(t, err)
				require.Empty(t, path)
				require.Contains(t, err.Error(), "failed to find executable")
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, path)
			}
		})
	}
}

func TestOSCommandExecutor_CommandContext(t *testing.T) {
	t.Parallel()

	executor := OSCommandExecutor{}

	//nolint:containedctx
	tests := []struct {
		name string
		ctx  context.Context
		cmd  string
		args []string
	}{
		{
			name: "basic command with context",
			ctx:  context.Background(),
			cmd:  "echo",
			args: []string{"hello"},
		},
		{
			name: "command with no args",
			ctx:  context.Background(),
			cmd:  "ls",
			args: nil,
		},
		{
			name: "command with multiple args",
			ctx:  context.Background(),
			cmd:  "echo",
			args: []string{"arg1", "arg2", "arg3"},
		},
		{
			name: "command with empty args",
			ctx:  context.Background(),
			cmd:  "pwd",
			args: []string{},
		},
		{
			name: "command with context timeout",
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 0)
				defer cancel()

				return ctx
			}(),
			cmd:  "sleep",
			args: []string{"1"},
		},
		{
			name: "command with cancelled context",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx
			}(),
			cmd:  "echo",
			args: []string{"test"},
		},
		{
			name: "command with deadline context",
			ctx: func() context.Context {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Hour))
				defer cancel()

				return ctx
			}(),
			cmd:  "echo",
			args: []string{"deadline"},
		},
		{
			name: "command with value context",
			ctx:  context.WithValue(context.Background(), testKey, "value"),
			cmd:  "echo",
			args: []string{"value"},
		},
		{
			name: "command with nil args slice",
			ctx:  context.Background(),
			cmd:  "true",
			args: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cmd := executor.CommandContext(testCase.ctx, testCase.cmd, testCase.args...)

			require.NotNil(t, cmd)
			require.Contains(t, cmd.Path(), testCase.cmd)
			require.Equal(t, append([]string{testCase.cmd}, testCase.args...), cmd.Args())
			// Note: exec.Cmd does not expose the context directly for comparison
			// The context is used internally by the command execution
		})
	}
}

func TestOSCommandExecutor_CommandContext_NilContext(t *testing.T) {
	t.Parallel()

	executor := OSCommandExecutor{}

	// Test with nil context - this should panic as per exec.CommandContext behavior
	require.Panics(t, func() {
		executor.CommandContext(nil, "echo", "test") //nolint:staticcheck // testing nil context causes panic
	})
}

func TestOSCommandExecutor_CommandContext_EdgeCases(t *testing.T) {
	t.Parallel()

	executor := OSCommandExecutor{}

	tests := []struct {
		name string
		cmd  string
		args []string
	}{
		{
			name: "empty command",
			cmd:  "",
			args: []string{},
		},
		{
			name: "command with special characters",
			cmd:  "echo",
			args: []string{"hello; rm -rf /", "world"},
		},
		{
			name: "command with very long args",
			cmd:  "echo",
			args: []string{string(make([]byte, 1000))},
		},
		{
			name: "command with unicode characters",
			cmd:  "echo",
			args: []string{"héllo", "wörld"},
		},
		{
			name: "command with control characters",
			cmd:  "echo",
			args: []string{"hello\x01world"},
		},
		{
			name: "command with extremely long command name",
			cmd:  string(make([]byte, 10000)),
			args: []string{},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cmd := executor.CommandContext(context.Background(), testCase.cmd, testCase.args...)

			assert.NotNil(t, cmd)
			assert.Contains(t, cmd.Path(), testCase.cmd)
			assert.Equal(t, append([]string{testCase.cmd}, testCase.args...), cmd.Args())
		})
	}
}

// TestOSCommandExecutor_LookPath_Success tests successful LookPath calls.
// Note: This test may fail in environments where common executables are not available.
func TestOSCommandExecutor_LookPath_Success(t *testing.T) {
	t.Parallel()

	executor := OSCommandExecutor{}

	// Test with common executables that should exist on most systems
	commonExecutables := []string{"sh", "bash", "true", "false"}

	for _, exe := range commonExecutables {
		t.Run("executable_"+exe, func(t *testing.T) {
			t.Parallel()

			path, err := executor.LookPath(exe)

			// If the executable exists, assert success
			if err == nil {
				require.NotEmpty(t, path)
				require.Contains(t, path, exe) // Path should contain the executable name
			} else {
				// If it doesn't exist, that's also acceptable for this test
				t.Logf("Executable %s not found: %v", exe, err)
			}
		})
	}
}

// TestOSCommandExecutor_CommandContext_TimeoutScenarios tests timeout and cancellation scenarios.
func TestOSCommandExecutor_CommandContext_TimeoutScenarios(t *testing.T) {
	t.Parallel()

	executor := OSCommandExecutor{}

	//nolint:containedctx
	tests := []struct {
		name        string
		ctx         context.Context
		cmd         string
		args        []string
		expectPanic bool
	}{
		{
			name: "context already timed out",
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
				cancel()
				time.Sleep(time.Millisecond) // Ensure timeout

				return ctx
			}(),
			cmd:         "sleep",
			args:        []string{"0.1"},
			expectPanic: false,
		},
		{
			name: "context with very short deadline",
			ctx: func() context.Context {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Nanosecond))
				defer cancel()

				return ctx
			}(),
			cmd:         "sleep",
			args:        []string{"0.1"},
			expectPanic: false,
		},
		{
			name:        "nil context should panic",
			ctx:         nil,
			cmd:         "echo",
			args:        []string{"test"},
			expectPanic: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if testCase.expectPanic {
				require.Panics(t, func() {
					executor.CommandContext(testCase.ctx, testCase.cmd, testCase.args...)
				})
			} else {
				cmd := executor.CommandContext(testCase.ctx, testCase.cmd, testCase.args...)
				require.NotNil(t, cmd)
				require.Contains(t, cmd.Path(), testCase.cmd)
				require.Equal(t, append([]string{testCase.cmd}, testCase.args...), cmd.Args())
			}
		})
	}
}

// TestOSCommandExecutor_CommandContext_PermissionIssues tests command execution with permission scenarios.
// Note: This test focuses on command creation, not actual execution.
func TestOSCommandExecutor_CommandContext_PermissionIssues(t *testing.T) {
	t.Parallel()

	executor := OSCommandExecutor{}

	tests := []struct {
		name string
		cmd  string
		args []string
	}{
		{
			name: "command requiring root permissions",
			cmd:  "mount",
			args: []string{"-t", "tmpfs", "tmpfs", "/tmp/test"},
		},
		{
			name: "command with setuid bit",
			cmd:  "sudo",
			args: []string{"whoami"},
		},
		{
			name: "command in restricted directory",
			cmd:  "/sbin/ifconfig",
			args: []string{},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cmd := executor.CommandContext(context.Background(), testCase.cmd, testCase.args...)

			require.NotNil(t, cmd)
			require.Contains(t, cmd.Path(), testCase.cmd)
			require.Equal(t, append([]string{testCase.cmd}, testCase.args...), cmd.Args())
		})
	}
}
