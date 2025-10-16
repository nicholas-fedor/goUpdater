// Package cli provides shared CLI display utilities for goUpdater.
// It includes functions for formatting output in a consistent tree-like structure,
// following modern CLI patterns similar to kubectl/docker.
package cli

import (
	"strings"
)

// TreeFormat formats a title and a slice of items into a tree-like structure.
// It provides a visual hierarchy with consistent alignment, using box-drawing characters
// for branches. The title is displayed first, followed by each item prefixed with
// tree branch symbols. The last item uses a corner branch (└─) while others use
// a tee branch (├─). Empty items are skipped to avoid unnecessary output.
//
// Parameters:
//   - title: The main title to display at the top of the tree
//   - items: A slice of strings representing the tree items to display
//
// Returns:
//   - A formatted string containing the tree structure
//
// Example:
//
//	result := TreeFormat("Go Installation Verification", []string{
//		"Directory: /usr/local/go",
//		"Version: go1.21.0",
//		"Status: verified",
//	})
//
// Output:
//
//	Go Installation Verification
//	├─ Directory: /usr/local/go
//	├─ Version: go1.21.0
//	└─ Status: verified
func TreeFormat(title string, items []string) string {
	var builder strings.Builder

	// Write the title
	builder.WriteString(title)
	builder.WriteString("\n")

	// Filter out empty items
	var filteredItems []string

	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			filteredItems = append(filteredItems, item)
		}
	}

	// Format each item with tree branches
	for i, item := range filteredItems {
		if i == len(filteredItems)-1 {
			// Last item uses corner branch
			builder.WriteString("└─ ")
		} else {
			// Other items use tee branch
			builder.WriteString("├─ ")
		}

		builder.WriteString(item)
		builder.WriteString("\n")
	}

	return builder.String()
}
