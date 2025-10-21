// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package privileges

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// OSPrivilegeManagerImpl provides OS-level privilege management operations.
// It implements the OSPrivilegeManager interface using standard library functions.
type OSPrivilegeManagerImpl struct{}

// Geteuid returns the effective user ID of the calling process.
// It wraps syscall.Geteuid() with proper error handling.
func (o OSPrivilegeManagerImpl) Geteuid() int {
	return syscall.Geteuid()
}

// Getuid returns the real user ID of the calling process.
// It wraps syscall.Getuid() with proper error handling.
func (o OSPrivilegeManagerImpl) Getuid() int {
	return syscall.Getuid()
}

// Getgid returns the real group ID of the calling process.
// It wraps syscall.Getgid() with proper error handling.
func (o OSPrivilegeManagerImpl) Getgid() int {
	return syscall.Getgid()
}

// Setuid sets the effective user ID of the calling process.
// It wraps syscall.Setuid() with proper error handling and context.
func (o OSPrivilegeManagerImpl) Setuid(uid int) error {
	err := syscall.Setuid(uid)
	if err != nil {
		return fmt.Errorf("failed to set UID to %d: %w", uid, err)
	}

	return nil
}

// Setgid sets the effective group ID of the calling process.
// It wraps syscall.Setgid() with proper error handling and context.
func (o OSPrivilegeManagerImpl) Setgid(gid int) error {
	err := syscall.Setgid(gid)
	if err != nil {
		return fmt.Errorf("failed to set GID to %d: %w", gid, err)
	}

	return nil
}

// Executable returns the path name for the executable that started the current process.
// It wraps os.Executable() with proper error handling.
func (o OSPrivilegeManagerImpl) Executable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	return exe, nil
}

// EvalSymlinks returns the path name after the evaluation of any symbolic links.
// It wraps filepath.EvalSymlinks() with proper error handling and context.
func (o OSPrivilegeManagerImpl) EvalSymlinks(path string) (string, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate symlinks for %q: %w", path, err)
	}

	return resolved, nil
}

// Exec replaces the current process with an execution of the named program with the given arguments.
// It wraps syscall.Exec() with proper error handling and context.
func (o OSPrivilegeManagerImpl) Exec(argv0 string, argv []string, envv []string) error {
	err := syscall.Exec(argv0, argv, envv)
	if err != nil {
		return fmt.Errorf("failed to execute %q: %w", argv0, err)
	}

	return nil
}

// Environ returns a copy of strings representing the environment, in the form "KEY=value".
// It wraps os.Environ() to provide the current environment variables.
func (o OSPrivilegeManagerImpl) Environ() []string {
	return os.Environ()
}

// Getenv retrieves the value of the environment variable named by the key.
// It wraps os.Getenv() to provide environment variable access.
func (o OSPrivilegeManagerImpl) Getenv(key string) string {
	return os.Getenv(key)
}
