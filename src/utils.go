package src

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strings"
)

// ExpandTildePrefix expands ~ and ~/ in file paths to the user's home directory.
// Returns an error if the current user cannot be determined.
func ExpandTildePrefix(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user for path expansion: %w", err)
	}
	homeDir := usr.HomeDir

	if path == "~" {
		return homeDir, nil
	}

	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDir, path[2:]), nil
	}

	// Path starts with ~ but not ~/ (e.g., ~username) - not supported
	return path, nil
}

// MustExpandTildePrefix expands ~ in file paths, panicking on error.
// Use this only when you're certain the operation will succeed.
func MustExpandTildePrefix(path string) string {
	expanded, err := ExpandTildePrefix(path)
	if err != nil {
		panic(err)
	}
	return expanded
}
