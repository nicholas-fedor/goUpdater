// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package version

import "errors"

var (
	// ErrVersionParseError indicates an error parsing the version.
	ErrVersionParseError = errors.New("version parse error")

	// ErrFailedToCreateEncoder is returned when creating a JSON encoder fails.
	ErrFailedToCreateEncoder = errors.New("failed to create JSON encoder: encoder is nil")
)
