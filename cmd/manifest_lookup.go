package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cageis/parts/src"
)

const manifestFilename = ".parts.yaml"

// resolveManifestPath returns the path to the manifest file.
// It checks the current directory first, then falls back to ~/.parts.yaml.
// When running under sudo, it also checks the invoking user's home directory.
func resolveManifestPath() string {
	// Check current directory
	if _, err := os.Stat(manifestFilename); err == nil {
		return manifestFilename
	}

	// Fall back to home directory
	homePath, err := src.ExpandTildePrefix("~/" + manifestFilename)
	if err == nil {
		if _, err := os.Stat(homePath); err == nil {
			return homePath
		}
	}

	// When running under sudo, check the invoking user's home directory
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		if sudoHome, err := sudoUserHome(sudoUser); err == nil {
			sudoPath := fmt.Sprintf("%s/%s", sudoHome, manifestFilename)
			if _, err := os.Stat(sudoPath); err == nil {
				return sudoPath
			}
		}
	}

	// Return default (will produce a clear error from LoadManifest)
	return manifestFilename
}

// sudoUserHome returns the home directory for the given username.
func sudoUserHome(username string) (string, error) {
	out, err := exec.Command("sh", "-c", fmt.Sprintf("eval echo ~%s", username)).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
