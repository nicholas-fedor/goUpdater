// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package validation

const (
	// MaxPathLength defines the maximum allowed length for file paths.
	MaxPathLength = 4096 // Common filesystem limit

	// MaxVersionLength defines the maximum allowed length for version strings.
	MaxVersionLength = 256

	// minVersionParts defines the minimum number of parts in a version string.
	minVersionParts = 2
)
