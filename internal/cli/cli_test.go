// Copyright © 2025 Nicholas Fedor
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"strings"
	"testing"
)

func TestTreeFormat(t *testing.T) {
	t.Parallel()
	testNormalCase(t)
	testSingleItem(t)
	testEmptyItemsSlice(t)
	testItemsWithEmptyStrings(t)
	testItemsWithWhitespaceOnly(t)
	testItemsWithLeadingTrailingSpaces(t)
	testEmptyTitle(t)
	testTitleWithSpecialCharacters(t)
	testAllItemsFilteredOut(t)
}

func testNormalCase(t *testing.T) {
	t.Helper()
	t.Run("normal case with multiple items", func(t *testing.T) {
		t.Parallel()

		result := TreeFormat("Go Installation Verification", []string{
			"Directory: /usr/local/go",
			"Version: go1.21.0",
			"Status: verified",
		})

		expected := "Go Installation Verification\n├─ Directory: /usr/local/go\n├─ Version: go1.21.0\n└─ Status: verified\n"
		if result != expected {
			t.Errorf("TreeFormat() = %q, want %q", result, expected)
		}
	})
}

func testSingleItem(t *testing.T) {
	t.Helper()
	t.Run("single item", func(t *testing.T) {
		t.Parallel()

		result := TreeFormat("Single Item", []string{"Only one"})

		expected := "Single Item\n└─ Only one\n"
		if result != expected {
			t.Errorf("TreeFormat() = %q, want %q", result, expected)
		}
	})
}

func testEmptyItemsSlice(t *testing.T) {
	t.Helper()
	t.Run("empty items slice", func(t *testing.T) {
		t.Parallel()

		result := TreeFormat("Empty List", []string{})

		expected := "Empty List\n"
		if result != expected {
			t.Errorf("TreeFormat() = %q, want %q", result, expected)
		}
	})
}

func testItemsWithEmptyStrings(t *testing.T) {
	t.Helper()
	t.Run("items with empty strings", func(t *testing.T) {
		t.Parallel()

		result := TreeFormat("Filtered Items", []string{"", "valid", "", "another"})

		expected := "Filtered Items\n├─ valid\n└─ another\n"
		if result != expected {
			t.Errorf("TreeFormat() = %q, want %q", result, expected)
		}
	})
}

func testItemsWithWhitespaceOnly(t *testing.T) {
	t.Helper()
	t.Run("items with whitespace only", func(t *testing.T) {
		t.Parallel()

		result := TreeFormat("Whitespace Items", []string{"   ", "\t", "valid item", "  "})

		expected := "Whitespace Items\n└─ valid item\n"
		if result != expected {
			t.Errorf("TreeFormat() = %q, want %q", result, expected)
		}
	})
}

func testItemsWithLeadingTrailingSpaces(t *testing.T) {
	t.Helper()
	t.Run("items with leading and trailing spaces", func(t *testing.T) {
		t.Parallel()

		result := TreeFormat("Spaced Items", []string{"  item1  ", "item2"})

		expected := "Spaced Items\n├─   item1  \n└─ item2\n"
		if result != expected {
			t.Errorf("TreeFormat() = %q, want %q", result, expected)
		}
	})
}

func testEmptyTitle(t *testing.T) {
	t.Helper()
	t.Run("empty title", func(t *testing.T) {
		t.Parallel()

		result := TreeFormat("", []string{"item1", "item2"})

		expected := "\n├─ item1\n└─ item2\n"
		if result != expected {
			t.Errorf("TreeFormat() = %q, want %q", result, expected)
		}
	})
}

func testTitleWithSpecialCharacters(t *testing.T) {
	t.Helper()
	t.Run("title with special characters", func(t *testing.T) {
		t.Parallel()

		result := TreeFormat("Title with émojis 🚀", []string{"item with unicode: ñ"})

		expected := "Title with émojis 🚀\n└─ item with unicode: ñ\n"
		if result != expected {
			t.Errorf("TreeFormat() = %q, want %q", result, expected)
		}
	})
}

func testAllItemsFilteredOut(t *testing.T) {
	t.Helper()
	t.Run("all items filtered out", func(t *testing.T) {
		t.Parallel()

		result := TreeFormat("All Filtered", []string{"", "   ", "\t\n"})

		expected := "All Filtered\n"
		if result != expected {
			t.Errorf("TreeFormat() = %q, want %q", result, expected)
		}
	})
}

// TestTreeFormatOutputFormat verifies the exact output format and structure.
func TestTreeFormatOutputFormat(t *testing.T) {
	t.Parallel()

	title := "Test Title"
	items := []string{"first", "second", "third"}

	result := TreeFormat(title, items)
	lines := strings.Split(strings.TrimSuffix(result, "\n"), "\n")

	// Check title is first line
	if lines[0] != title {
		t.Errorf("First line should be title: got %q, want %q", lines[0], title)
	}

	// Check number of lines: title + number of items
	expectedLines := 1 + len(items)
	if len(lines) != expectedLines {
		t.Errorf("Expected %d lines, got %d", expectedLines, len(lines))
	}

	// Check branch symbols
	expectedBranches := []string{"├─ ", "├─ ", "└─ "}
	for index, branch := range expectedBranches {
		line := lines[index+1]
		if !strings.HasPrefix(line, branch) {
			t.Errorf("Line %d should start with %q, got %q", index+1, branch, line)
		}

		if strings.TrimPrefix(line, branch) != items[index] {
			t.Errorf("Line %d content after branch should be %q, got %q",
				index+1, items[index], strings.TrimPrefix(line, branch))
		}
	}
}

// TestTreeFormatFiltering verifies that empty and whitespace-only items are filtered correctly.
func TestTreeFormatFiltering(t *testing.T) {
	t.Parallel()

	items := []string{"", "valid", "   ", "another", "\t", "last valid"}

	result := TreeFormat("Test", items)
	lines := strings.Split(strings.TrimSuffix(result, "\n"), "\n")

	// Should have title + 3 valid items
	if len(lines) != 4 {
		t.Errorf("Expected 4 lines (title + 3 items), got %d", len(lines))
	}

	// Check that only valid items appear
	expectedItems := []string{"valid", "another", "last valid"}
	for itemIndex, expected := range expectedItems {
		line := lines[itemIndex+1]

		content := strings.TrimPrefix(line, "├─ ")
		if itemIndex == len(expectedItems)-1 {
			content = strings.TrimPrefix(line, "└─ ")
		}

		if content != expected {
			t.Errorf("Expected item %q, got %q", expected, content)
		}
	}
}
