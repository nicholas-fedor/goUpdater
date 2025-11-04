// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package privileges

import (
	"strconv"

	"github.com/nicholas-fedor/goUpdater/internal/logger"
)

// LogPrivilegeChange logs privilege changes for audit purposes.
// It records the operation type, user IDs involved, and the reason for the change.
func (d *DefaultAuditLogger) LogPrivilegeChange(operation string, fromUID, toUID int, reason string) {
	logger.Infof("Privilege change: %s from UID %d to UID %d - %s", operation, fromUID, toUID, reason)
}

// LogElevationAttempt logs attempts to elevate privileges.
// It differentiates between successful and failed elevation attempts with appropriate log levels.
func (d *DefaultAuditLogger) LogElevationAttempt(success bool, reason string) {
	if success {
		logger.Debug("Privilege elevation attempt initiated: " + reason)
	} else {
		logger.Warn("Privilege elevation attempt failed: " + reason)
	}
}

// LogPrivilegeDrop logs attempts to drop privileges.
// It records the target user ID and differentiates between successful and failed drop attempts.
func (d *DefaultAuditLogger) LogPrivilegeDrop(success bool, targetUID int, reason string) {
	if success {
		logger.Info("Privilege drop attempt to UID " + strconv.Itoa(targetUID) + ": " + reason)
	} else {
		logger.Warn("Privilege drop attempt to UID " + strconv.Itoa(targetUID) + " failed: " + reason)
	}
}
