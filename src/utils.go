package src

import (
	"os/user"
	"path/filepath"
	"strings"
)

// ExpandTildePrefix expands ~ and ~/ in file paths to the user's home directory
func ExpandTildePrefix(path string) string {
	usr, _ := user.Current()
	homeDir := usr.HomeDir

	if path == "~" {
		return homeDir
	}

	if strings.HasPrefix(path, "~/") {
		path = filepath.Join(homeDir, path[2:])
	}

	return path
}
