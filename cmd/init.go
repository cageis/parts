package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cageis/parts/src"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const manifestTemplate = `# Parts manifest — manages dotfiles from this directory
# Docs: https://github.com/cageis/parts

# Default settings applied to all targets (can be overridden per-target)
defaults:
  comment: "auto"    # auto-detect comment style from file extension
  backup: false      # create .bak files before modifying targets
  # mode: merge      # 'merge' (default) or 'own'

# Each target defines a file to manage
targets:
  # Example: merge SSH config partials into ~/.ssh/config
  # ssh:
  #   target: ~/.ssh/config
  #   partials: ./ssh/
  #   comment: "#"
  #   mode: merge      # preserves content outside PARTIALS markers

  # Example: fully manage ~/.vimrc from partials
  # vimrc:
  #   target: ~/.vimrc
  #   partials: ./vim/
  #   mode: own         # entire file is written from partials
`

// initManifestPath allows tests to override the manifest location
var initManifestPath string

func newInitCmd() *cobra.Command {
	var fromArgs []string
	var targetName string
	var targetMode string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Generate a .parts.yaml manifest",
		Long: `Creates a .parts.yaml file in the current directory.

Without flags, generates a skeleton manifest with commented examples.

With --from, creates (or appends to) a manifest from existing CLI arguments:
  parts init --from <target-file> <partials-dir> <comment-style>`,
		Example: `  # Generate skeleton manifest
  parts init

  # Create manifest from existing usage
  parts init --from ~/.ssh/config ./ssh "#" --name ssh

  # Append another target to existing manifest
  parts init --from ~/.vimrc ./vim "auto" --name vimrc --mode own`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(fromArgs) > 0 {
				return runInitFrom(fromArgs, targetName, targetMode)
			}
			return runInitSkeleton()
		},
	}

	cmd.Flags().StringArrayVar(&fromArgs, "from", nil, "migrate from CLI args: <target-file> <partials-dir> <comment-style>")
	cmd.Flags().StringVar(&targetName, "name", "", "target name (derived from filename if omitted)")
	cmd.Flags().StringVar(&targetMode, "mode", "merge", "target mode: merge or own")

	return cmd
}

func runInitSkeleton() error {
	if _, err := os.Stat(".parts.yaml"); err == nil {
		return fmt.Errorf(".parts.yaml already exists in this directory")
	}

	if err := os.WriteFile(".parts.yaml", []byte(manifestTemplate), 0644); err != nil {
		return fmt.Errorf("failed to create .parts.yaml: %w", err)
	}

	fmt.Println("Created .parts.yaml — edit it to define your targets")
	return nil
}

func runInitFrom(fromArgs []string, name, mode string) error {
	if len(fromArgs) != 3 {
		return fmt.Errorf("--from requires exactly 3 values: <target-file> <partials-dir> <comment-style>, got %d", len(fromArgs))
	}

	targetFile := fromArgs[0]
	partialsDir := fromArgs[1]
	commentStyle := fromArgs[2]

	// Validate mode
	if mode != "merge" && mode != "own" {
		return fmt.Errorf("invalid mode '%s': must be 'merge' or 'own'", mode)
	}

	// Validate partials directory exists
	expandedPartials, err := src.ExpandTildePrefix(partialsDir)
	if err != nil {
		return fmt.Errorf("failed to expand partials path: %w", err)
	}
	info, err := os.Stat(expandedPartials)
	if err != nil {
		return fmt.Errorf("partials directory '%s' does not exist", partialsDir)
	}
	if !info.IsDir() {
		return fmt.Errorf("'%s' is not a directory", partialsDir)
	}

	// Normalize paths: convert relative paths to absolute, then shorten $HOME to ~/
	partialsDir = normalizePath(partialsDir, expandedPartials)
	targetFile = normalizeTargetPath(targetFile)

	// Derive name if not provided
	if name == "" {
		name = deriveTargetName(targetFile)
	}

	manifestPath := initManifestPath
	if manifestPath == "" {
		manifestPath = resolveManifestPath()
	}

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		// No manifest found anywhere — create in cwd
		return createManifestWithTarget(manifestPath, name, targetFile, partialsDir, commentStyle, mode)
	}

	// Append to existing manifest (cwd or ~/.parts.yaml)
	return appendTargetToManifest(manifestPath, name, targetFile, partialsDir, commentStyle, mode)
}

// deriveTargetName generates a target name from a file path.
// ~/.ssh/config -> ssh-config, ~/.vimrc -> vimrc, /etc/hosts -> hosts
func deriveTargetName(path string) string {
	// Remove tilde prefix for derivation
	clean := strings.TrimPrefix(path, "~/")
	clean = strings.TrimPrefix(clean, "/")

	// Get the last two path components for context
	parts := strings.Split(clean, "/")
	var name string
	if len(parts) >= 2 {
		// Use parent-file pattern: .ssh/config -> ssh-config
		parent := strings.TrimPrefix(parts[len(parts)-2], ".")
		file := strings.TrimPrefix(parts[len(parts)-1], ".")
		if parent != "" && file != "" {
			name = parent + "-" + file
		} else if file != "" {
			name = file
		} else {
			name = parent
		}
	} else {
		name = strings.TrimPrefix(filepath.Base(clean), ".")
	}

	// Clean up: lowercase, replace special chars
	name = strings.ToLower(name)
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, name)
	name = strings.Trim(name, "-")

	if name == "" {
		name = "target"
	}
	return name
}

func createManifestWithTarget(path, name, target, partials, comment, mode string) error {
	manifest := struct {
		Defaults map[string]string            `yaml:"defaults"`
		Targets  map[string]map[string]string `yaml:"targets"`
	}{
		Defaults: map[string]string{
			"comment": "auto",
		},
		Targets: map[string]map[string]string{
			name: {
				"target":   target,
				"partials": partials,
				"comment":  comment,
				"mode":     mode,
			},
		},
	}

	data, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	fmt.Printf("Created %s with target '%s'\n", path, name)
	return nil
}

// normalizePath converts a path to use ~/ when it falls under the user's home directory.
// It takes the original user-provided path and the already-expanded absolute path.
func normalizePath(original, expanded string) string {
	// Already uses tilde — leave as-is
	if strings.HasPrefix(original, "~/") || original == "~" {
		return original
	}

	// Resolve to absolute if relative
	abs := expanded
	if !filepath.IsAbs(abs) {
		if resolved, err := filepath.Abs(abs); err == nil {
			abs = resolved
		}
	}

	// Shorten to ~/ if under home directory
	usr, err := os.UserHomeDir()
	if err != nil {
		return original
	}

	if strings.HasPrefix(abs, usr+"/") {
		return "~/" + strings.TrimPrefix(abs, usr+"/")
	}
	if abs == usr {
		return "~"
	}

	return abs
}

// normalizeTargetPath converts a target file path to use ~/ when under $HOME.
func normalizeTargetPath(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		return path
	}

	abs := path
	if !filepath.IsAbs(abs) {
		if resolved, err := filepath.Abs(abs); err == nil {
			abs = resolved
		}
	}

	usr, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if strings.HasPrefix(abs, usr+"/") {
		return "~/" + strings.TrimPrefix(abs, usr+"/")
	}

	return abs
}

func appendTargetToManifest(path, name, target, partials, comment, mode string) error {
	// Load existing manifest to check for duplicates
	existing, err := src.LoadManifest(path)
	if err != nil {
		return fmt.Errorf("failed to read existing manifest: %w", err)
	}

	if _, exists := existing.Targets[name]; exists {
		return fmt.Errorf("target '%s' already exists in manifest", name)
	}

	// Read raw YAML, unmarshal into generic structure, add target, re-marshal
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	targets, ok := raw["targets"].(map[string]interface{})
	if !ok {
		targets = make(map[string]interface{})
	}

	targets[name] = map[string]interface{}{
		"target":   target,
		"partials": partials,
		"comment":  comment,
		"mode":     mode,
	}
	raw["targets"] = targets

	out, err := yaml.Marshal(raw)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(path, out, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	fmt.Printf("Added target '%s' to %s\n", name, path)
	return nil
}
