package src

import (
	"os/user"
	"path/filepath"
	"strings"
	"testing"
)

func TestExpandTildePrefix(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}
	homeDir := usr.HomeDir

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Just tilde",
			input:    "~",
			expected: homeDir,
		},
		{
			name:     "Tilde with path",
			input:    "~/Documents/test.txt",
			expected: filepath.Join(homeDir, "Documents/test.txt"),
		},
		{
			name:     "Absolute path unchanged",
			input:    "/var/www/test.txt",
			expected: "/var/www/test.txt",
		},
		{
			name:     "Relative path unchanged",
			input:    "relative/path.txt",
			expected: "relative/path.txt",
		},
		{
			name:     "Empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "Tilde in middle unchanged",
			input:    "/home/~user/file",
			expected: "/home/~user/file",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ExpandTildePrefix(test.input)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, result)
			}
		})
	}
}

func TestExpandTildePrefix_UnsupportedFormat(t *testing.T) {
	// ~username format is passed through unchanged (not supported)
	result, err := ExpandTildePrefix("~otheruser/path")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// Should return as-is since ~username format is not supported
	if result != "~otheruser/path" {
		t.Errorf("Expected ~otheruser/path to be unchanged, got %q", result)
	}
}

func TestMustExpandTildePrefix(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Fatalf("Failed to get current user: %v", err)
	}

	// Test that it works for valid paths
	result := MustExpandTildePrefix("~/test")
	expected := filepath.Join(usr.HomeDir, "test")
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}

	// Test non-tilde path passes through
	result = MustExpandTildePrefix("/absolute/path")
	if result != "/absolute/path" {
		t.Errorf("Expected /absolute/path, got %q", result)
	}
}

func TestConstants(t *testing.T) {
	// Verify constants have expected values
	if PartialStartMarker != "PARTIALS>>>>>" {
		t.Errorf("PartialStartMarker has unexpected value: %q", PartialStartMarker)
	}
	if PartialEndMarker != "PARTIALS<<<<<" {
		t.Errorf("PartialEndMarker has unexpected value: %q", PartialEndMarker)
	}
	if MarkerSeparator != "============================" {
		t.Errorf("MarkerSeparator has unexpected value: %q", MarkerSeparator)
	}

	// Verify markers contain the separator
	// This is a sanity check that constants work together
	if !strings.Contains("# "+MarkerSeparator, "====") {
		t.Error("MarkerSeparator should contain equals signs")
	}
}
