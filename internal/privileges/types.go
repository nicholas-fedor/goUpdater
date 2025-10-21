// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package privileges

// OSPrivilegeManager handles OS-level privilege operations.
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
}

// AuditLogger provides logging functionality for privilege operations.
type AuditLogger interface {
	LogPrivilegeChange(operation string, fromUID, toUID int, reason string)
	LogElevationAttempt(success bool, reason string)
	LogPrivilegeDrop(success bool, targetUID int, reason string)
}
