package src

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// ExpandTildePrefix expands ~ and ~/ in file paths to the user's home directory.
// When running under sudo, it uses the invoking user's home directory (via SUDO_USER).
// Returns an error if the current user cannot be determined.
func ExpandTildePrefix(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	homeDir, err := resolveHomeDir()
	if err != nil {
		return "", err
	}

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

// resolveHomeDir returns the home directory of the invoking user.
// When running under sudo, it looks up the SUDO_USER's home directory
// so that ~ expands to the real user's home, not root's.
func resolveHomeDir() (string, error) {
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		u, err := user.Lookup(sudoUser)
		if err == nil {
			return u.HomeDir, nil
		}
	}

	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("failed to get current user for path expansion: %w", err)
	}
	return usr.HomeDir, nil
}
