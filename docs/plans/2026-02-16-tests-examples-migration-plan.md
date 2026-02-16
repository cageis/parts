# Tests, Examples & Migration — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add `init --from` migration feature, fill test coverage gaps for manifest workflows, and create runnable examples for the new features.

**Architecture:** Extend the existing `init` subcommand with `--from` flags. Add new test files for migration, end-to-end workflows, and edge cases. Create three example directories following the established pattern in `examples/`.

**Tech Stack:** Go 1.17, cobra CLI, gopkg.in/yaml.v3, standard `testing` package.

---

### Task 1: Add `init --from` core logic

**Files:**
- Modify: `cmd/init.go`

**Step 1: Write the `initFrom` function and wire `--from` flag into the existing init command**

Replace the contents of `cmd/init.go` with:

```go
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

	// Derive name if not provided
	if name == "" {
		name = deriveTargetName(targetFile)
	}

	manifestPath := ".parts.yaml"

	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		// Create new manifest
		return createManifestWithTarget(manifestPath, name, targetFile, partialsDir, commentStyle, mode)
	}

	// Append to existing manifest
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
```

**Step 2: Run existing tests to verify skeleton init still works**

Run: `go test ./cmd/ -run TestInit -v`
Expected: PASS — `TestInitCommand_CreatesManifest` and `TestInitCommand_DoesNotOverwrite` both pass.

**Step 3: Commit**

```bash
git add cmd/init.go
git commit -m "feat(init): add --from flag for legacy CLI migration"
```

---

### Task 2: Write `init --from` tests

**Files:**
- Create: `cmd/init_from_test.go`

**Step 1: Write all `init --from` tests**

```go
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cageis/parts/src"
)

func TestInitFrom_CreatesNewManifest(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	// Create partials directory with a partial file
	partialsDir := filepath.Join(dir, "ssh")
	os.MkdirAll(partialsDir, 0755)
	os.WriteFile(filepath.Join(partialsDir, "work"), []byte("Host work\n    User admin\n"), 0644)

	// Create target file
	targetFile := filepath.Join(dir, "config")
	os.WriteFile(targetFile, []byte("# My SSH config\n"), 0644)

	cmd := newInitCmd()
	cmd.SetArgs([]string{"--from", targetFile, "--from", "./ssh", "--from", "#", "--name", "ssh"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init --from failed: %v", err)
	}

	// Verify manifest was created
	content, err := os.ReadFile(".parts.yaml")
	if err != nil {
		t.Fatalf("Manifest not created: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "ssh") {
		t.Error("Manifest should contain target name 'ssh'")
	}
	if !strings.Contains(contentStr, targetFile) {
		t.Error("Manifest should contain target file path")
	}
	if !strings.Contains(contentStr, "./ssh") {
		t.Error("Manifest should contain partials directory")
	}
}

func TestInitFrom_AppendsToExisting(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	// Create initial manifest with one target
	partialsDir1 := filepath.Join(dir, "ssh")
	os.MkdirAll(partialsDir1, 0755)
	os.WriteFile(filepath.Join(partialsDir1, "work"), []byte("Host work\n"), 0644)

	targetFile1 := filepath.Join(dir, "ssh-config")
	os.WriteFile(targetFile1, []byte("# SSH\n"), 0644)

	cmd1 := newInitCmd()
	cmd1.SetArgs([]string{"--from", targetFile1, "--from", "./ssh", "--from", "#", "--name", "ssh"})
	if err := cmd1.Execute(); err != nil {
		t.Fatalf("First init --from failed: %v", err)
	}

	// Add second target
	partialsDir2 := filepath.Join(dir, "vim")
	os.MkdirAll(partialsDir2, 0755)
	os.WriteFile(filepath.Join(partialsDir2, "config"), []byte("set number\n"), 0644)

	targetFile2 := filepath.Join(dir, "vimrc")
	os.WriteFile(targetFile2, []byte("\" vimrc\n"), 0644)

	cmd2 := newInitCmd()
	cmd2.SetArgs([]string{"--from", targetFile2, "--from", "./vim", "--from", "\"", "--name", "vimrc", "--mode", "own"})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("Second init --from failed: %v", err)
	}

	// Verify manifest has both targets
	content, _ := os.ReadFile(".parts.yaml")
	contentStr := string(content)
	if !strings.Contains(contentStr, "ssh") {
		t.Error("Manifest should still contain 'ssh' target")
	}
	if !strings.Contains(contentStr, "vimrc") {
		t.Error("Manifest should contain 'vimrc' target")
	}
}

func TestInitFrom_DeriveNameFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"~/.ssh/config", "ssh-config"},
		{"~/.vimrc", "vimrc"},
		{"/etc/hosts", "hosts"},
		{"~/.gitconfig", "gitconfig"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := deriveTargetName(tt.path)
			if got != tt.expected {
				t.Errorf("deriveTargetName(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}

func TestInitFrom_ExplicitName(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	partialsDir := filepath.Join(dir, "partials")
	os.MkdirAll(partialsDir, 0755)
	os.WriteFile(filepath.Join(partialsDir, "part1"), []byte("content\n"), 0644)

	cmd := newInitCmd()
	cmd.SetArgs([]string{"--from", "/tmp/myfile", "--from", "./partials", "--from", "#", "--name", "custom-name"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("init --from failed: %v", err)
	}

	content, _ := os.ReadFile(".parts.yaml")
	if !strings.Contains(string(content), "custom-name") {
		t.Error("Manifest should use the explicit name 'custom-name'")
	}
}

func TestInitFrom_ValidatesPartialsDir(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	cmd := newInitCmd()
	cmd.SetArgs([]string{"--from", "/tmp/target", "--from", "./nonexistent", "--from", "#"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error for missing partials directory")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected 'does not exist' error, got: %v", err)
	}
}

func TestInitFrom_DuplicateTargetName(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	partialsDir := filepath.Join(dir, "ssh")
	os.MkdirAll(partialsDir, 0755)
	os.WriteFile(filepath.Join(partialsDir, "work"), []byte("Host work\n"), 0644)

	// Create first target
	cmd1 := newInitCmd()
	cmd1.SetArgs([]string{"--from", "/tmp/config", "--from", "./ssh", "--from", "#", "--name", "ssh"})
	if err := cmd1.Execute(); err != nil {
		t.Fatalf("First init failed: %v", err)
	}

	// Try duplicate
	cmd2 := newInitCmd()
	cmd2.SetArgs([]string{"--from", "/tmp/other", "--from", "./ssh", "--from", "#", "--name", "ssh"})
	err := cmd2.Execute()
	if err == nil {
		t.Fatal("Expected error for duplicate target name")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Expected 'already exists' error, got: %v", err)
	}
}

func TestInitFrom_GeneratedManifestIsUsable(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	// Create partials
	partialsDir := filepath.Join(dir, "ssh")
	os.MkdirAll(partialsDir, 0755)
	os.WriteFile(filepath.Join(partialsDir, "work"), []byte("Host work\n    User admin\n"), 0644)

	// Create target file
	targetFile := filepath.Join(dir, "config")
	os.WriteFile(targetFile, []byte("# My SSH config\n"), 0644)

	// init --from
	initCmd := newInitCmd()
	initCmd.SetArgs([]string{"--from", targetFile, "--from", partialsDir, "--from", "#", "--name", "ssh"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init --from failed: %v", err)
	}

	// Verify the generated manifest can be loaded and used
	manifest, err := src.LoadManifest(filepath.Join(dir, ".parts.yaml"))
	if err != nil {
		t.Fatalf("Generated manifest is not loadable: %v", err)
	}

	if len(manifest.Targets) != 1 {
		t.Fatalf("Expected 1 target, got %d", len(manifest.Targets))
	}

	target := manifest.ResolvedTarget("ssh")
	if target.Target != targetFile {
		t.Errorf("Expected target path %q, got %q", targetFile, target.Target)
	}

	// Apply the manifest target to verify it actually works
	buildCmd, buildErr := src.NewPartialsBuildCommand(target.Target, target.Partials, target.Comment)
	if buildErr != nil {
		t.Fatalf("Failed to create build command: %v", buildErr)
	}
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Apply failed on generated manifest: %v", err)
	}

	// Verify the target file was updated
	result, _ := os.ReadFile(targetFile)
	if !strings.Contains(string(result), "Host work") {
		t.Error("Target file should contain merged partial content")
	}
}
```

**Step 2: Run the new tests**

Run: `go test ./cmd/ -run TestInitFrom -v`
Expected: All 7 tests PASS.

**Step 3: Run all tests to verify nothing broke**

Run: `go test ./... -v`
Expected: All tests PASS (existing + new).

**Step 4: Commit**

```bash
git add cmd/init_from_test.go
git commit -m "test(init): add comprehensive init --from migration tests"
```

---

### Task 3: Add `init` skeleton validation test

**Files:**
- Modify: `cmd/init_test.go`

**Step 1: Add `TestInitCommand_GeneratedManifestIsParseable` test**

Append to the end of `cmd/init_test.go` (before closing):

```go
func TestInitCommand_GeneratedManifestIsParseable(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	cmd := newInitCmd()
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// The skeleton YAML should be parseable (even though all targets are commented out,
	// it should parse without errors — it just won't have valid targets)
	content, _ := os.ReadFile(filepath.Join(dir, ".parts.yaml"))
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(content, &parsed); err != nil {
		t.Fatalf("Generated skeleton is not valid YAML: %v", err)
	}

	if _, ok := parsed["defaults"]; !ok {
		t.Error("Skeleton should have a 'defaults' key")
	}
}
```

Also add to the imports at the top of `cmd/init_test.go`:

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)
```

**Step 2: Run the test**

Run: `go test ./cmd/ -run TestInitCommand_Generated -v`
Expected: PASS.

**Step 3: Commit**

```bash
git add cmd/init_test.go
git commit -m "test(init): verify skeleton manifest is parseable YAML"
```

---

### Task 4: Add sync edge case tests

**Files:**
- Modify: `src/sync_test.go`

**Step 1: Add 5 new sync tests**

Append to the end of `src/sync_test.go`:

```go
func TestExtractPartialSections_BlockComments(t *testing.T) {
	content := `/* Source: /tmp/partials/reset.css */
* { margin: 0; }
/* Source: /tmp/partials/layout.css */
.container { width: 100%; }
`
	sections, err := ExtractPartialSections(content, "/*")
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	if sections["/tmp/partials/reset.css"] != "* { margin: 0; }\n" {
		t.Errorf("Unexpected reset content: %q", sections["/tmp/partials/reset.css"])
	}
	if sections["/tmp/partials/layout.css"] != ".container { width: 100%; }\n" {
		t.Errorf("Unexpected layout content: %q", sections["/tmp/partials/layout.css"])
	}
}

func TestExtractPartialSections_HTMLComments(t *testing.T) {
	content := `<!-- Source: /tmp/partials/header.html -->
<header>My Site</header>
<!-- Source: /tmp/partials/nav.html -->
<nav>Home | About</nav>
`
	sections, err := ExtractPartialSections(content, "<!--")
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	if sections["/tmp/partials/header.html"] != "<header>My Site</header>\n" {
		t.Errorf("Unexpected header content: %q", sections["/tmp/partials/header.html"])
	}
	if sections["/tmp/partials/nav.html"] != "<nav>Home | About</nav>\n" {
		t.Errorf("Unexpected nav content: %q", sections["/tmp/partials/nav.html"])
	}
}

func TestSyncTarget_OwnMode(t *testing.T) {
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	os.MkdirAll(partialsDir, 0755)

	os.WriteFile(filepath.Join(partialsDir, "header"), []byte("#!/bin/bash\nset -e\n"), 0644)
	os.WriteFile(filepath.Join(partialsDir, "body"), []byte("echo hello\n"), 0644)

	// Create target using own mode
	targetFile := filepath.Join(dir, "script.sh")
	ownCmd := NewPartialsOwnCommand(targetFile, partialsDir, "#")
	if err := ownCmd.Run(); err != nil {
		t.Fatalf("Own command failed: %v", err)
	}

	// Modify the target
	content, _ := os.ReadFile(targetFile)
	modified := strings.Replace(string(content), "echo hello", "echo world", 1)
	os.WriteFile(targetFile, []byte(modified), 0644)

	// Sync in own mode
	result, err := SyncTarget(targetFile, partialsDir, "#", "own", false)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if result.UpdatedFiles != 1 {
		t.Errorf("Expected 1 updated file, got %d", result.UpdatedFiles)
	}

	// Verify the partial was updated
	bodyContent, _ := os.ReadFile(filepath.Join(partialsDir, "body"))
	if !strings.Contains(string(bodyContent), "echo world") {
		t.Error("Partial should contain updated content 'echo world'")
	}
}

func TestSyncTarget_MissingTargetFile(t *testing.T) {
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	os.MkdirAll(partialsDir, 0755)

	_, err := SyncTarget(filepath.Join(dir, "nonexistent"), partialsDir, "#", "merge", false)
	if err == nil {
		t.Fatal("Expected error for missing target file")
	}
	if !strings.Contains(err.Error(), "failed to read target file") {
		t.Errorf("Expected 'failed to read' error, got: %v", err)
	}
}

func TestSyncTarget_NoManagedSection(t *testing.T) {
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	os.MkdirAll(partialsDir, 0755)

	// Target with no PARTIALS markers
	targetFile := filepath.Join(dir, "config")
	os.WriteFile(targetFile, []byte("# Just a plain config\nHost work\n"), 0644)

	result, err := SyncTarget(targetFile, partialsDir, "#", "merge", false)
	if err != nil {
		t.Fatalf("Sync should not error on file without markers: %v", err)
	}

	if result.UpdatedFiles != 0 {
		t.Errorf("Expected 0 updates when no managed section, got %d", result.UpdatedFiles)
	}
}
```

**Step 2: Run the new sync tests**

Run: `go test ./src/ -run "TestExtractPartialSections_Block|TestExtractPartialSections_HTML|TestSyncTarget_Own|TestSyncTarget_Missing|TestSyncTarget_NoManaged" -v`
Expected: All 5 tests PASS.

**Step 3: Commit**

```bash
git add src/sync_test.go
git commit -m "test(sync): add block comments, HTML, own mode, and edge case tests"
```

---

### Task 5: Add own mode edge case tests

**Files:**
- Modify: `src/own_test.go`

**Step 1: Add 3 new own mode tests**

Append to the end of `src/own_test.go`:

```go
func TestPartialsOwnCommand_EmptyPartialsDir(t *testing.T) {
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	os.MkdirAll(partialsDir, 0755)
	targetFile := filepath.Join(dir, "target")

	cmd := NewPartialsOwnCommand(targetFile, partialsDir, "")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	result, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target: %v", err)
	}

	if string(result) != "" {
		t.Errorf("Expected empty file for empty partials dir, got: %q", string(result))
	}
}

func TestPartialsOwnCommand_MultiplePartialsOrdering(t *testing.T) {
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	os.MkdirAll(partialsDir, 0755)
	targetFile := filepath.Join(dir, "target")

	// Create 5 partials with intentionally unordered names
	names := []string{"05-last", "01-first", "03-middle", "02-second", "04-fourth"}
	for _, name := range names {
		os.WriteFile(filepath.Join(partialsDir, name), []byte(name+"\n"), 0644)
	}

	cmd := NewPartialsOwnCommand(targetFile, partialsDir, "")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	result, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target: %v", err)
	}

	// os.ReadDir returns alphabetical order
	expected := "01-first\n02-second\n03-middle\n04-fourth\n05-last\n"
	if string(result) != expected {
		t.Errorf("Expected alphabetical ordering:\n%q\nGot:\n%q", expected, string(result))
	}
}

func TestPartialsOwnCommand_NoTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	os.MkdirAll(partialsDir, 0755)
	targetFile := filepath.Join(dir, "target")

	// Write partial WITHOUT trailing newline
	os.WriteFile(filepath.Join(partialsDir, "part1"), []byte("no trailing newline"), 0644)

	cmd := NewPartialsOwnCommand(targetFile, partialsDir, "")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	result, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target: %v", err)
	}

	// Should normalize to have a trailing newline
	if string(result) != "no trailing newline\n" {
		t.Errorf("Expected normalized trailing newline, got: %q", string(result))
	}
}
```

**Step 2: Run the new own mode tests**

Run: `go test ./src/ -run "TestPartialsOwnCommand_Empty|TestPartialsOwnCommand_Multiple|TestPartialsOwnCommand_NoTrailing" -v`
Expected: All 3 tests PASS.

**Step 3: Commit**

```bash
git add src/own_test.go
git commit -m "test(own): add empty dir, ordering, and newline normalization tests"
```

---

### Task 6: Add manifest edge case tests

**Files:**
- Modify: `src/manifest_test.go`

**Step 1: Add 3 new manifest tests**

Append to the end of `src/manifest_test.go` (before the helper functions):

```go
func TestLoadManifest_TildePaths(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, ".parts.yaml")

	yaml := `targets:
  ssh:
    target: ~/. ssh/config
    partials: ~/dotfiles/ssh/
`
	// Note: this tests that the YAML parser correctly reads tilde paths as strings.
	// Actual tilde expansion happens at apply time, not parse time.
	if err := os.WriteFile(manifestPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest with tilde paths: %v", err)
	}

	ssh := manifest.Targets["ssh"]
	if ssh.Partials != "~/dotfiles/ssh/" {
		t.Errorf("Expected tilde path preserved, got: %q", ssh.Partials)
	}
}

func TestLoadManifest_EmptyTargetMap(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, ".parts.yaml")

	yaml := `defaults:
  comment: "#"
targets:
`
	if err := os.WriteFile(manifestPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	_, err := LoadManifest(manifestPath)
	if err == nil {
		t.Fatal("Expected error for empty targets section")
	}
	if !containsSubstring(err.Error(), "no targets defined") {
		t.Errorf("Expected 'no targets defined' error, got: %v", err)
	}
}

func TestLoadManifest_SpecialTargetNames(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, ".parts.yaml")

	yaml := `targets:
  ssh-config:
    target: /tmp/ssh
    partials: ./ssh/
  vim_rc:
    target: /tmp/vim
    partials: ./vim/
  git.config:
    target: /tmp/git
    partials: ./git/
`
	if err := os.WriteFile(manifestPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest with special target names: %v", err)
	}

	expected := []string{"ssh-config", "vim_rc", "git.config"}
	for _, name := range expected {
		if _, ok := manifest.Targets[name]; !ok {
			t.Errorf("Expected target '%s' to be parsed", name)
		}
	}
}
```

**Step 2: Run the new manifest tests**

Run: `go test ./src/ -run "TestLoadManifest_Tilde|TestLoadManifest_Empty|TestLoadManifest_Special" -v`
Expected: All 3 tests PASS.

**Step 3: Commit**

```bash
git add src/manifest_test.go
git commit -m "test(manifest): add tilde paths, empty targets, and special names tests"
```

---

### Task 7: Write end-to-end workflow tests

**Files:**
- Create: `cmd/workflow_test.go`

**Step 1: Write all 4 workflow tests**

```go
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cageis/parts/src"
)

func TestWorkflow_InitFromApplyRemoveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	// Set up partials
	partialsDir := filepath.Join(dir, "ssh")
	os.MkdirAll(partialsDir, 0755)
	os.WriteFile(filepath.Join(partialsDir, "work"), []byte("Host work\n    User admin\n"), 0644)

	// Set up target
	targetFile := filepath.Join(dir, "config")
	os.WriteFile(targetFile, []byte("# My SSH config\nHost personal\n    User me\n"), 0644)

	// Step 1: init --from
	initCmd := newInitCmd()
	initCmd.SetArgs([]string{"--from", targetFile, "--from", partialsDir, "--from", "#", "--name", "ssh"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init --from failed: %v", err)
	}

	// Override manifest path for apply/remove commands
	manifestPath := filepath.Join(dir, ".parts.yaml")
	applyManifestPath = manifestPath
	manifestRemovePath = manifestPath
	defer func() {
		applyManifestPath = ""
		manifestRemovePath = ""
	}()

	// Step 2: apply
	applyCmd := newApplyCmd()
	applyCmd.SetArgs([]string{})
	if err := applyCmd.Execute(); err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	// Verify target has merged content
	content, _ := os.ReadFile(targetFile)
	if !strings.Contains(string(content), "Host work") {
		t.Error("After apply, target should contain partial content")
	}
	if !strings.Contains(string(content), "Host personal") {
		t.Error("After apply, target should preserve original content")
	}

	// Step 3: remove
	removeCmd := newManifestRemoveCmd()
	removeCmd.SetArgs([]string{})
	if err := removeCmd.Execute(); err != nil {
		t.Fatalf("remove failed: %v", err)
	}

	// Verify partials section is removed, original content preserved
	content, _ = os.ReadFile(targetFile)
	if strings.Contains(string(content), "Host work") {
		t.Error("After remove, target should not contain partial content")
	}
	if !strings.Contains(string(content), "Host personal") {
		t.Error("After remove, target should preserve original content")
	}
}

func TestWorkflow_ApplyEditSyncReapply(t *testing.T) {
	dir := t.TempDir()

	// Set up
	partialsDir := filepath.Join(dir, "ssh")
	os.MkdirAll(partialsDir, 0755)
	os.WriteFile(filepath.Join(partialsDir, "work"), []byte("Host work\n    User admin\n    Port 22\n"), 0644)

	targetFile := filepath.Join(dir, "config")
	os.WriteFile(targetFile, []byte("# My config\n"), 0644)

	// Write manifest
	manifest := `targets:
  ssh:
    target: ` + targetFile + `
    partials: ` + partialsDir + `
    comment: "#"
    mode: merge
`
	manifestPath := filepath.Join(dir, ".parts.yaml")
	os.WriteFile(manifestPath, []byte(manifest), 0644)

	applyManifestPath = manifestPath
	syncManifestPath = manifestPath
	defer func() {
		applyManifestPath = ""
		syncManifestPath = ""
	}()

	// Step 1: apply
	applyCmd := newApplyCmd()
	applyCmd.SetArgs([]string{})
	if err := applyCmd.Execute(); err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	// Step 2: edit target (simulate user changing Port 22 to Port 2222)
	content, _ := os.ReadFile(targetFile)
	modified := strings.Replace(string(content), "Port 22", "Port 2222", 1)
	os.WriteFile(targetFile, []byte(modified), 0644)

	// Step 3: sync
	syncCmd := newSyncCmd()
	syncCmd.SetArgs([]string{})
	if err := syncCmd.Execute(); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	// Verify partial was updated
	partialContent, _ := os.ReadFile(filepath.Join(partialsDir, "work"))
	if !strings.Contains(string(partialContent), "Port 2222") {
		t.Error("After sync, partial should contain updated port")
	}

	// Step 4: re-apply (should produce same result as current target)
	applyCmd2 := newApplyCmd()
	applyCmd2.SetArgs([]string{})
	if err := applyCmd2.Execute(); err != nil {
		t.Fatalf("re-apply failed: %v", err)
	}

	// Verify target still has the change
	finalContent, _ := os.ReadFile(targetFile)
	if !strings.Contains(string(finalContent), "Port 2222") {
		t.Error("After re-apply, target should still have synced change")
	}
}

func TestWorkflow_MixedModesApplyAndRemove(t *testing.T) {
	dir := t.TempDir()

	// Set up 3 targets: ssh (merge), bashrc (own), gitconfig (own)
	sshPartials := filepath.Join(dir, "ssh")
	os.MkdirAll(sshPartials, 0755)
	os.WriteFile(filepath.Join(sshPartials, "work"), []byte("Host work\n    User admin\n"), 0644)

	bashPartials := filepath.Join(dir, "bash")
	os.MkdirAll(bashPartials, 0755)
	os.WriteFile(filepath.Join(bashPartials, "01-exports"), []byte("export PATH=/usr/bin\n"), 0644)
	os.WriteFile(filepath.Join(bashPartials, "02-aliases"), []byte("alias ll='ls -la'\n"), 0644)

	gitPartials := filepath.Join(dir, "git")
	os.MkdirAll(gitPartials, 0755)
	os.WriteFile(filepath.Join(gitPartials, "config"), []byte("[user]\n    name = Test\n"), 0644)

	sshTarget := filepath.Join(dir, "ssh-config")
	os.WriteFile(sshTarget, []byte("# SSH Config\nHost personal\n    User me\n"), 0644)

	bashTarget := filepath.Join(dir, "bashrc")
	gitTarget := filepath.Join(dir, "gitconfig")

	manifest := `targets:
  ssh:
    target: ` + sshTarget + `
    partials: ` + sshPartials + `
    comment: "#"
    mode: merge
  bashrc:
    target: ` + bashTarget + `
    partials: ` + bashPartials + `
    comment: "#"
    mode: own
  gitconfig:
    target: ` + gitTarget + `
    partials: ` + gitPartials + `
    mode: own
`
	manifestPath := filepath.Join(dir, ".parts.yaml")
	os.WriteFile(manifestPath, []byte(manifest), 0644)

	applyManifestPath = manifestPath
	manifestRemovePath = manifestPath
	defer func() {
		applyManifestPath = ""
		manifestRemovePath = ""
	}()

	// Apply all
	applyCmd := newApplyCmd()
	applyCmd.SetArgs([]string{})
	if err := applyCmd.Execute(); err != nil {
		t.Fatalf("apply all failed: %v", err)
	}

	// Verify all targets were created/updated
	sshContent, _ := os.ReadFile(sshTarget)
	if !strings.Contains(string(sshContent), "Host work") || !strings.Contains(string(sshContent), "Host personal") {
		t.Error("SSH target should have both merged and original content")
	}

	bashContent, _ := os.ReadFile(bashTarget)
	if !strings.Contains(string(bashContent), "export PATH") || !strings.Contains(string(bashContent), "alias ll") {
		t.Error("Bash target should have all partials concatenated")
	}

	gitContent, _ := os.ReadFile(gitTarget)
	if !strings.Contains(string(gitContent), "[user]") {
		t.Error("Git target should have git config content")
	}

	// Remove all
	removeCmd := newManifestRemoveCmd()
	removeCmd.SetArgs([]string{})
	if err := removeCmd.Execute(); err != nil {
		t.Fatalf("remove all failed: %v", err)
	}

	// Verify: merge target has partials removed, own targets are deleted
	sshContent, _ = os.ReadFile(sshTarget)
	if strings.Contains(string(sshContent), "Host work") {
		t.Error("After remove, SSH should not have partial content")
	}
	if !strings.Contains(string(sshContent), "Host personal") {
		t.Error("After remove, SSH should preserve original content")
	}

	if _, err := os.Stat(bashTarget); !os.IsNotExist(err) {
		t.Error("After remove, own-mode bashrc should be deleted")
	}
	if _, err := os.Stat(gitTarget); !os.IsNotExist(err) {
		t.Error("After remove, own-mode gitconfig should be deleted")
	}
}

func TestWorkflow_SelectiveOperations(t *testing.T) {
	dir := t.TempDir()

	// Two merge targets
	sshPartials := filepath.Join(dir, "ssh")
	os.MkdirAll(sshPartials, 0755)
	os.WriteFile(filepath.Join(sshPartials, "work"), []byte("Host work\n"), 0644)

	gitPartials := filepath.Join(dir, "git")
	os.MkdirAll(gitPartials, 0755)
	os.WriteFile(filepath.Join(gitPartials, "config"), []byte("[core]\n    editor = vim\n"), 0644)

	sshTarget := filepath.Join(dir, "ssh-config")
	os.WriteFile(sshTarget, []byte("# SSH\n"), 0644)

	gitTarget := filepath.Join(dir, "gitconfig")
	os.WriteFile(gitTarget, []byte("# Git\n"), 0644)

	manifest := `targets:
  ssh:
    target: ` + sshTarget + `
    partials: ` + sshPartials + `
    comment: "#"
    mode: merge
  git:
    target: ` + gitTarget + `
    partials: ` + gitPartials + `
    comment: "#"
    mode: merge
`
	manifestPath := filepath.Join(dir, ".parts.yaml")
	os.WriteFile(manifestPath, []byte(manifest), 0644)

	applyManifestPath = manifestPath
	syncManifestPath = manifestPath
	manifestRemovePath = manifestPath
	defer func() {
		applyManifestPath = ""
		syncManifestPath = ""
		manifestRemovePath = ""
	}()

	// Apply only ssh
	applyCmd := newApplyCmd()
	applyCmd.SetArgs([]string{"ssh"})
	if err := applyCmd.Execute(); err != nil {
		t.Fatalf("selective apply failed: %v", err)
	}

	// SSH should have partials, git should be untouched
	sshContent, _ := os.ReadFile(sshTarget)
	if !strings.Contains(string(sshContent), "Host work") {
		t.Error("SSH target should have partial content after selective apply")
	}

	gitContent, _ := os.ReadFile(gitTarget)
	if strings.Contains(string(gitContent), "editor") {
		t.Error("Git target should be untouched after selective SSH apply")
	}

	// Now apply git
	applyCmd2 := newApplyCmd()
	applyCmd2.SetArgs([]string{"git"})
	if err := applyCmd2.Execute(); err != nil {
		t.Fatalf("selective git apply failed: %v", err)
	}

	gitContent, _ = os.ReadFile(gitTarget)
	if !strings.Contains(string(gitContent), "editor") {
		t.Error("Git target should have partial content after selective apply")
	}

	// Remove only ssh, git should remain
	removeCmd := newManifestRemoveCmd()
	removeCmd.SetArgs([]string{"ssh"})
	if err := removeCmd.Execute(); err != nil {
		t.Fatalf("selective remove failed: %v", err)
	}

	sshContent, _ = os.ReadFile(sshTarget)
	if strings.Contains(string(sshContent), "Host work") {
		t.Error("SSH should be cleaned after selective remove")
	}

	gitContent, _ = os.ReadFile(gitTarget)
	if !strings.Contains(string(gitContent), "editor") {
		t.Error("Git should be untouched after selective SSH remove")
	}
}
```

**Step 2: Run the workflow tests**

Run: `go test ./cmd/ -run TestWorkflow -v`
Expected: All 4 tests PASS.

**Step 3: Run all tests to verify nothing broke**

Run: `go test ./...`
Expected: All tests PASS.

**Step 4: Commit**

```bash
git add cmd/workflow_test.go
git commit -m "test(workflow): add end-to-end integration tests for manifest workflows"
```

---

### Task 8: Create `examples/manifest-dotfiles/` example

**Files:**
- Create: `examples/manifest-dotfiles/.parts.yaml`
- Create: `examples/manifest-dotfiles/ssh/work`
- Create: `examples/manifest-dotfiles/ssh/personal`
- Create: `examples/manifest-dotfiles/bash/01-exports`
- Create: `examples/manifest-dotfiles/bash/02-aliases`
- Create: `examples/manifest-dotfiles/git/gitconfig`
- Create: `examples/manifest-dotfiles/run-example.sh`
- Create: `examples/manifest-dotfiles/README.md`

**Step 1: Create directory structure and files**

Create `examples/manifest-dotfiles/.parts.yaml`:
```yaml
defaults:
  comment: "auto"

targets:
  ssh:
    target: ./output/ssh-config
    partials: ./ssh/
    comment: "#"
    mode: merge
  bashrc:
    target: ./output/bashrc
    partials: ./bash/
    comment: "#"
    mode: own
  gitconfig:
    target: ./output/gitconfig
    partials: ./git/
    mode: own
```

Create `examples/manifest-dotfiles/ssh/work`:
```
Host work
    HostName work.example.com
    User deploy
    Port 22
    IdentityFile ~/.ssh/id_work
```

Create `examples/manifest-dotfiles/ssh/personal`:
```
Host personal
    HostName personal.example.com
    User me
    Port 22
    IdentityFile ~/.ssh/id_personal
```

Create `examples/manifest-dotfiles/bash/01-exports`:
```
export EDITOR=vim
export PATH="$HOME/bin:$PATH"
export LANG=en_US.UTF-8
```

Create `examples/manifest-dotfiles/bash/02-aliases`:
```
alias ll='ls -la'
alias gs='git status'
alias gp='git push'
```

Create `examples/manifest-dotfiles/git/gitconfig`:
```
[user]
    name = Example User
    email = user@example.com
[core]
    editor = vim
    autocrlf = input
[pull]
    rebase = true
```

Create `examples/manifest-dotfiles/run-example.sh`:
```bash
#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PARTS="$SCRIPT_DIR/../../bin/parts"
cd "$SCRIPT_DIR"

echo "=== Manifest Dotfiles Example ==="
echo "Demonstrates multi-target manifest with merge + own modes."
echo

# Create base SSH config (merge mode preserves this)
mkdir -p output
echo "# My SSH Config" > output/ssh-config
echo "# Custom settings above will be preserved" >> output/ssh-config
echo "" >> output/ssh-config

echo "--- Step 1: Apply all targets ---"
$PARTS apply
echo

echo "SSH config (merge mode - original content preserved):"
cat output/ssh-config
echo

echo "Bashrc (own mode - entirely from partials):"
cat output/bashrc
echo

echo "Gitconfig (own mode):"
cat output/gitconfig
echo

echo "--- Step 2: Selective apply (ssh only) ---"
# First remove to reset
$PARTS remove
echo

# Re-create base SSH config
echo "# My SSH Config" > output/ssh-config
echo "# Custom settings above will be preserved" >> output/ssh-config
echo "" >> output/ssh-config

$PARTS apply ssh
echo

echo "SSH config updated:"
cat output/ssh-config
echo

echo "Bashrc should not exist yet:"
ls output/bashrc 2>/dev/null && echo "(exists)" || echo "(not created - correct)"
echo

echo "--- Step 3: Apply remaining targets ---"
$PARTS apply bashrc gitconfig
echo

echo "--- Step 4: Sync demo (edit target, sync back) ---"
# Simulate user editing the SSH config
sed -i.bak 's/Port 22/Port 2222/g' output/ssh-config 2>/dev/null || \
    sed 's/Port 22/Port 2222/g' output/ssh-config > output/ssh-config.tmp && mv output/ssh-config.tmp output/ssh-config
echo "Edited SSH config (changed Port 22 -> 2222)"
$PARTS sync ssh
echo "Partial updated:"
cat ssh/work
echo

echo "--- Step 5: Remove all ---"
$PARTS remove
echo

echo "SSH config after remove (original content preserved):"
cat output/ssh-config
echo

echo "Bashrc after remove (deleted):"
ls output/bashrc 2>/dev/null && echo "(still exists - error)" || echo "(deleted - correct)"
echo

# Cleanup
rm -rf output
rm -f ssh/work.bak

echo "=== Example completed ==="
```

Create `examples/manifest-dotfiles/README.md`:
```markdown
# Manifest Dotfiles Example

Demonstrates managing multiple dotfiles using a `.parts.yaml` manifest with both **merge** and **own** modes.

## Targets

| Name | Mode | Description |
|------|------|-------------|
| `ssh` | merge | SSH config — partials merged between markers, user content preserved |
| `bashrc` | own | Bashrc — entire file written from concatenated partials |
| `gitconfig` | own | Git config — entire file from partials |

## Running

```bash
# Build parts first
cd ../.. && make build && cd examples/manifest-dotfiles

# Run the example
./run-example.sh
```

## What It Demonstrates

1. **Apply all** — all 3 targets applied in one command
2. **Selective apply** — apply only specific targets
3. **Sync** — edit target file, sync changes back to partials
4. **Remove** — clean up: merge targets strip markers, own targets delete file
```

**Step 2: Make the script executable and test it**

Run: `chmod +x examples/manifest-dotfiles/run-example.sh`
Run: `cd /private/var/www/sandbox/parts && make build && cd examples/manifest-dotfiles && ./run-example.sh`
Expected: Script runs without errors, demonstrating all workflow steps.

**Step 3: Commit**

```bash
git add examples/manifest-dotfiles/
git commit -m "docs(examples): add manifest-dotfiles multi-target example"
```

---

### Task 9: Create `examples/manifest-sync-workflow/` example

**Files:**
- Create: `examples/manifest-sync-workflow/.parts.yaml`
- Create: `examples/manifest-sync-workflow/ssh/work`
- Create: `examples/manifest-sync-workflow/ssh/staging`
- Create: `examples/manifest-sync-workflow/run-example.sh`
- Create: `examples/manifest-sync-workflow/README.md`

**Step 1: Create directory structure and files**

Create `examples/manifest-sync-workflow/.parts.yaml`:
```yaml
targets:
  ssh:
    target: ./output/ssh-config
    partials: ./ssh/
    comment: "#"
    mode: merge
```

Create `examples/manifest-sync-workflow/ssh/work`:
```
Host work
    HostName work.example.com
    User deploy
    Port 22
```

Create `examples/manifest-sync-workflow/ssh/staging`:
```
Host staging
    HostName staging.example.com
    User deploy
    Port 22
```

Create `examples/manifest-sync-workflow/run-example.sh`:
```bash
#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PARTS="$SCRIPT_DIR/../../bin/parts"
cd "$SCRIPT_DIR"

echo "=== Sync Workflow Example ==="
echo "Demonstrates bidirectional sync between target and partials."
echo

mkdir -p output
echo "# SSH Config - managed by parts" > output/ssh-config
echo "" >> output/ssh-config

echo "--- Step 1: Initial apply ---"
$PARTS apply
echo
echo "Target file:"
cat output/ssh-config
echo

echo "--- Step 2: Simulate editing the target file ---"
# Change port for work host
sed -i.bak 's/Port 22/Port 2222/' output/ssh-config 2>/dev/null || \
    sed 's/Port 22/Port 2222/' output/ssh-config > output/ssh-config.tmp && mv output/ssh-config.tmp output/ssh-config
echo "Changed all Port 22 -> Port 2222 in target"
echo
echo "Modified target:"
cat output/ssh-config
echo

echo "--- Step 3: Sync changes back to partials ---"
$PARTS sync
echo
echo "Work partial after sync:"
cat ssh/work
echo
echo "Staging partial after sync:"
cat ssh/staging
echo

echo "--- Step 4: Re-apply to verify round-trip ---"
$PARTS apply
echo
echo "Target after re-apply (should match step 2 edit):"
cat output/ssh-config
echo

echo "--- Step 5: Cleanup ---"
$PARTS remove
rm -rf output

# Restore original partials
echo 'Host work
    HostName work.example.com
    User deploy
    Port 22' > ssh/work
echo 'Host staging
    HostName staging.example.com
    User deploy
    Port 22' > ssh/staging

echo
echo "=== Sync workflow example completed ==="
```

Create `examples/manifest-sync-workflow/README.md`:
```markdown
# Manifest Sync Workflow Example

Demonstrates bidirectional sync between a target file and its partial source files.

## Workflow

1. **Apply** — merge partials into the target
2. **Edit** — make changes directly in the target file
3. **Sync** — pull changes back into the individual partial files
4. **Re-apply** — verify the round-trip preserves changes

## Running

```bash
cd ../.. && make build && cd examples/manifest-sync-workflow
./run-example.sh
```

## Key Concept

The sync feature uses `# Source: <path>` comments to map sections of the target file back to their source partial files. When you edit the target and run `parts sync`, the changes flow back to the correct partial.
```

**Step 2: Make the script executable and test it**

Run: `chmod +x examples/manifest-sync-workflow/run-example.sh`
Run: `cd /private/var/www/sandbox/parts && cd examples/manifest-sync-workflow && ./run-example.sh`
Expected: Script runs without errors.

**Step 3: Commit**

```bash
git add examples/manifest-sync-workflow/
git commit -m "docs(examples): add manifest-sync-workflow bidirectional sync example"
```

---

### Task 10: Create `examples/manifest-migration/` example

**Files:**
- Create: `examples/manifest-migration/partials/work.conf`
- Create: `examples/manifest-migration/run-example.sh`
- Create: `examples/manifest-migration/README.md`

**Step 1: Create directory structure and files**

Create `examples/manifest-migration/partials/work.conf`:
```
Host work
    HostName work.example.com
    User deploy
    Port 22
    IdentityFile ~/.ssh/id_work
```

Create `examples/manifest-migration/run-example.sh`:
```bash
#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PARTS="$SCRIPT_DIR/../../bin/parts"
cd "$SCRIPT_DIR"

# Clean previous runs
rm -f .parts.yaml
rm -rf output

echo "=== Migration Example ==="
echo "Demonstrates migrating from legacy CLI to manifest mode."
echo

mkdir -p output
echo "# SSH Config" > output/ssh-config
echo "" >> output/ssh-config

echo "--- Step 1: Legacy CLI usage ---"
echo "Running: parts output/ssh-config partials \"#\""
$PARTS output/ssh-config partials "#"
echo
echo "Result (legacy):"
cat output/ssh-config
echo

echo "--- Step 2: Remove legacy partials ---"
$PARTS --remove output/ssh-config "#"
echo "Cleaned target file."
echo

echo "--- Step 3: Migrate to manifest with init --from ---"
echo "Running: parts init --from output/ssh-config --from ./partials --from \"#\" --name ssh"
$PARTS init --from output/ssh-config --from ./partials --from "#" --name ssh
echo
echo "Generated .parts.yaml:"
cat .parts.yaml
echo

echo "--- Step 4: Apply using manifest ---"
$PARTS apply
echo
echo "Result (manifest):"
cat output/ssh-config
echo

echo "--- Step 5: Cleanup ---"
$PARTS remove
rm -f .parts.yaml
rm -rf output

echo
echo "=== Migration example completed ==="
```

Create `examples/manifest-migration/README.md`:
```markdown
# Migration Example

Demonstrates migrating from the legacy positional-argument CLI to the new manifest-driven workflow.

## Legacy vs Manifest

**Legacy (positional args):**
```bash
parts ~/.ssh/config ~/.ssh/config.d "#"
```

**Manifest (`.parts.yaml`):**
```bash
parts init --from ~/.ssh/config --from ./ssh --from "#" --name ssh
parts apply
```

## Running

```bash
cd ../.. && make build && cd examples/manifest-migration
./run-example.sh
```

## Migration Steps

1. Run your existing legacy command one last time
2. Remove the partials section: `parts --remove <file> <comment>`
3. Create manifest: `parts init --from <file> --from <dir> --from <comment> --name <name>`
4. Apply with manifest: `parts apply`
```

**Step 2: Make the script executable and test it**

Run: `chmod +x examples/manifest-migration/run-example.sh`
Run: `cd /private/var/www/sandbox/parts && cd examples/manifest-migration && ./run-example.sh`
Expected: Script runs without errors.

**Step 3: Commit**

```bash
git add examples/manifest-migration/
git commit -m "docs(examples): add manifest-migration legacy-to-manifest example"
```

---

### Task 11: Update examples README and run-all script

**Files:**
- Modify: `examples/README.md`
- Modify: `examples/run-all-examples.sh`

**Step 1: Add new examples to `examples/README.md`**

Add the following section after the existing directory listing (after `- `python/`...`):

```markdown
- `manifest-dotfiles/` - Multi-target manifest with merge + own modes
- `manifest-sync-workflow/` - Bidirectional sync between target and partials
- `manifest-migration/` - Migration from legacy CLI to manifest mode
```

**Step 2: Add new examples to `examples/run-all-examples.sh`**

Add before the final success message:

```bash
echo
echo "📁 Manifest Dotfiles Example"
echo "============================"
cd manifest-dotfiles
./run-example.sh
cd ..

echo
echo "📁 Manifest Sync Workflow Example"
echo "================================="
cd manifest-sync-workflow
./run-example.sh
cd ..

echo
echo "📁 Manifest Migration Example"
echo "=============================="
cd manifest-migration
./run-example.sh
cd ..
```

**Step 3: Run all tests to verify nothing broke**

Run: `go test ./...`
Expected: All tests PASS.

**Step 4: Commit**

```bash
git add examples/README.md examples/run-all-examples.sh
git commit -m "docs(examples): add manifest examples to README and run-all script"
```

---

### Task 12: Final verification

**Step 1: Run all tests**

Run: `go test ./... -v`
Expected: All tests PASS. Should include the new tests:
- `TestInitFrom_*` (7 tests)
- `TestInitCommand_GeneratedManifestIsParseable`
- `TestWorkflow_*` (4 tests)
- `TestExtractPartialSections_BlockComments`
- `TestExtractPartialSections_HTMLComments`
- `TestSyncTarget_OwnMode`
- `TestSyncTarget_MissingTargetFile`
- `TestSyncTarget_NoManagedSection`
- `TestPartialsOwnCommand_EmptyPartialsDir`
- `TestPartialsOwnCommand_MultiplePartialsOrdering`
- `TestPartialsOwnCommand_NoTrailingNewline`
- `TestLoadManifest_TildePaths`
- `TestLoadManifest_EmptyTargetMap`
- `TestLoadManifest_SpecialTargetNames`

**Step 2: Build and run all examples**

Run: `make build && cd examples && ./run-all-examples.sh`
Expected: All examples complete successfully.

**Step 3: Count test coverage**

Run: `go test ./... -v 2>&1 | grep -c "=== RUN"`
Expected: 60+ total tests (existing 48 + 22 new).
