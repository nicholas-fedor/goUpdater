// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package privileges

import (
	"strconv"

	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

// dropPrivileges drops privileges back to the original user after performing elevated operations.
// It only attempts to drop privileges if the process is currently running as root and was elevated via sudo.
// The method follows the security practice of dropping group privileges before user privileges.
//
//nolint:funlen
func (p *PrivilegeManager) dropPrivileges() error {
	// Only drop privileges if we are currently running as root and were elevated via sudo
	if !p.isRoot() || !isElevated(p.pm) {
		return nil // Not elevated, nothing to drop
	}

	// Get the original user's UID and GID from SUDO_UID and SUDO_GID
	sudoUID := p.pm.Getenv("SUDO_UID")
	sudoGID := p.pm.Getenv("SUDO_GID")

	logger.Debugf("Parsed SUDO_UID='%s', SUDO_GID='%s'", sudoUID, sudoGID)

	if sudoUID == "" || sudoGID == "" {
		return &PrivilegeDropError{
			TargetUID: 0, // We don't know the target
			Reason:    "SUDO_UID or SUDO_GID environment variables not set",
			Cause:     nil,
		}
	}

	// Parse the UIDs/GIDs
	originalUID, err := strconv.Atoi(sudoUID)
	if err != nil {
		return &PrivilegeDropError{
			TargetUID: 0,
			Reason:    "invalid SUDO_UID value",
			Cause:     err,
		}
	}

	originalGID, err := strconv.Atoi(sudoGID)
	if err != nil {
		return &PrivilegeDropError{
			TargetUID: originalUID,
			Reason:    "invalid SUDO_GID value",
			Cause:     err,
		}
	}

	logger.Debugf("Parsed UID=%d, GID=%d", originalUID, originalGID)

	logger.Debugf("Calling LogPrivilegeDrop with success=true, uid=%d, message='dropping privileges to original user'", //nolint:lll
		originalUID)
	// Log the privilege drop attempt
	p.logger.LogPrivilegeDrop(true, originalUID, "dropping privileges to original user")

	logger.Debugf("Calling Setgid with gid=%d", originalGID)
	// Drop group privileges first (recommended practice)
	err = p.pm.Setgid(originalGID)
	if err != nil {
		dropErr := &PrivilegeDropError{
			TargetUID: originalUID,
			Reason:    "failed to set GID",
			Cause:     err,
		}
		p.logger.LogPrivilegeDrop(false, originalUID, "failed to drop group privileges")

		return dropErr
	}

	logger.Debugf("Calling Setuid with uid=%d", originalUID)
	// Then drop user privileges
	err = p.pm.Setuid(originalUID)
	if err != nil {
		dropErr := &PrivilegeDropError{
			TargetUID: originalUID,
			Reason:    "failed to set UID",
			Cause:     err,
		}
		p.logger.LogPrivilegeDrop(false, originalUID, "failed to drop user privileges")

		return dropErr
	}

	logger.Debug("Privilege drop completed successfully")

	// Log successful privilege drop
	p.logger.LogPrivilegeChange("drop", 0, originalUID, "dropped privileges to original user")

	return nil
}
