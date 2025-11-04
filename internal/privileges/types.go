// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package privileges

import (
	"github.com/nicholas-fedor/goUpdater/internal/exec"
	"github.com/nicholas-fedor/goUpdater/internal/filesystem"
)

// OSPrivilegeManager handles OS-level privilege operations.
//
//nolint:interfacebloat // Interface requires multiple OS-level methods for comprehensive privilege management
type OSPrivilegeManager interface {
	Geteuid() int
	Executable() (string, error)
	EvalSymlinks(path string) (string, error)
	Exec(argv0 string, argv []string, envv []string) error
	Environ() []string
	Setuid(uid int) error
	Setgid(gid int) error
	Getuid() int
	Getgid() int
	Getenv(key string) string
	Args() []string
}

// AuditLogger provides logging functionality for privilege operations.
type AuditLogger interface {
	LogPrivilegeChange(operation string, fromUID, toUID int, reason string)
	LogElevationAttempt(success bool, reason string)
	LogPrivilegeDrop(success bool, targetUID int, reason string)
}

// DefaultAuditLogger provides a basic implementation of audit logging using the logger package.
type DefaultAuditLogger struct{}

// OSPrivilegeManagerImpl provides OS-level privilege management operations.
// It implements the OSPrivilegeManager interface using standard library functions.
type OSPrivilegeManagerImpl struct{}

// PrivilegeManager handles privilege escalation with dependency injection.
type PrivilegeManager struct {
	fs       filesystem.FileSystem
	pm       OSPrivilegeManager
	executor exec.CommandExecutor
	logger   AuditLogger
}
