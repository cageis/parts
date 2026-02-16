# Manifest-Driven Dotfile Management — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add manifest-driven dotfile management to Parts, supporting multi-target apply/remove/sync via `.parts.yaml`, with full backward compatibility for the existing positional-arg CLI.

**Architecture:** New cobra subcommands (`apply`, `remove`, `sync`, `init`) read a `.parts.yaml` manifest and orchestrate the existing `PartialsBuildCommand`/`PartialsRemoveCommand` engines for `merge` mode targets, plus a new `own` mode engine for whole-file targets. Existing CLI behavior is completely untouched — subcommands vs positional args are unambiguous to cobra.

**Tech Stack:** Go 1.17+, cobra (already used), `gopkg.in/yaml.v3` (new dependency)

---

## Task 1: Add YAML dependency

**Files:**
- Modify: `go.mod`

**Step 1: Add yaml.v3 dependency**

Run: `cd /private/var/www/sandbox/parts && go get gopkg.in/yaml.v3`

**Step 2: Verify it was added**

Run: `grep yaml go.mod`
Expected: line containing `gopkg.in/yaml.v3`

**Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add gopkg.in/yaml.v3 dependency"
```

---

## Task 2: Manifest parsing and validation

**Files:**
- Create: `src/manifest.go`
- Create: `src/manifest_test.go`

**Step 1: Write the failing test for manifest parsing**

Create `src/manifest_test.go`:

```go
package src

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifest_Basic(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, ".parts.yaml")

	yaml := `defaults:
  comment: "auto"
  backup: true

targets:
  ssh:
    target: /tmp/test-ssh-config
    partials: ./ssh/
    comment: "#"
    mode: merge
  vimrc:
    target: /tmp/test-vimrc
    partials: ./vim/
    mode: own
`
	if err := os.WriteFile(manifestPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	if manifest.Defaults.Comment != "auto" {
		t.Errorf("Expected default comment 'auto', got '%s'", manifest.Defaults.Comment)
	}
	if manifest.Defaults.Backup != true {
		t.Error("Expected default backup true")
	}
	if len(manifest.Targets) != 2 {
		t.Fatalf("Expected 2 targets, got %d", len(manifest.Targets))
	}

	ssh := manifest.Targets["ssh"]
	if ssh.Target != "/tmp/test-ssh-config" {
		t.Errorf("Expected ssh target '/tmp/test-ssh-config', got '%s'", ssh.Target)
	}
	if ssh.Mode != "merge" {
		t.Errorf("Expected ssh mode 'merge', got '%s'", ssh.Mode)
	}
	if ssh.Comment != "#" {
		t.Errorf("Expected ssh comment '#', got '%s'", ssh.Comment)
	}

	vim := manifest.Targets["vimrc"]
	if vim.Mode != "own" {
		t.Errorf("Expected vimrc mode 'own', got '%s'", vim.Mode)
	}
}

func TestLoadManifest_DefaultsInheritance(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, ".parts.yaml")

	yaml := `defaults:
  comment: "#"
  mode: merge

targets:
  ssh:
    target: /tmp/ssh-config
    partials: ./ssh/
  git:
    target: /tmp/gitconfig
    partials: ./git/
    comment: "//"
    mode: own
`
	if err := os.WriteFile(manifestPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	// ssh should inherit defaults
	ssh := manifest.ResolvedTarget("ssh")
	if ssh.Comment != "#" {
		t.Errorf("Expected ssh to inherit comment '#', got '%s'", ssh.Comment)
	}
	if ssh.Mode != "merge" {
		t.Errorf("Expected ssh to inherit mode 'merge', got '%s'", ssh.Mode)
	}

	// git should use its own overrides
	git := manifest.ResolvedTarget("git")
	if git.Comment != "//" {
		t.Errorf("Expected git comment '//', got '%s'", git.Comment)
	}
	if git.Mode != "own" {
		t.Errorf("Expected git mode 'own', got '%s'", git.Mode)
	}
}

func TestLoadManifest_BuiltinDefaults(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, ".parts.yaml")

	// No defaults section at all
	yaml := `targets:
  ssh:
    target: /tmp/ssh-config
    partials: ./ssh/
`
	if err := os.WriteFile(manifestPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	ssh := manifest.ResolvedTarget("ssh")
	if ssh.Comment != "auto" {
		t.Errorf("Expected builtin default comment 'auto', got '%s'", ssh.Comment)
	}
	if ssh.Mode != "merge" {
		t.Errorf("Expected builtin default mode 'merge', got '%s'", ssh.Mode)
	}
}

func TestLoadManifest_Validation(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name:    "missing targets",
			yaml:    `defaults: {comment: "#"}`,
			wantErr: "no targets defined",
		},
		{
			name: "missing target path",
			yaml: `targets:
  ssh:
    partials: ./ssh/
`,
			wantErr: "target 'ssh': missing 'target' path",
		},
		{
			name: "missing partials path",
			yaml: `targets:
  ssh:
    target: /tmp/config
`,
			wantErr: "target 'ssh': missing 'partials' path",
		},
		{
			name: "invalid mode",
			yaml: `targets:
  ssh:
    target: /tmp/config
    partials: ./ssh/
    mode: symlink
`,
			wantErr: "target 'ssh': invalid mode 'symlink'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifestPath := filepath.Join(dir, tt.name+".yaml")
			if err := os.WriteFile(manifestPath, []byte(tt.yaml), 0644); err != nil {
				t.Fatalf("Failed to write manifest: %v", err)
			}

			_, err := LoadManifest(manifestPath)
			if err == nil {
				t.Fatal("Expected validation error, got nil")
			}
			if !containsString(err.Error(), tt.wantErr) {
				t.Errorf("Expected error containing '%s', got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestLoadManifest_FileNotFound(t *testing.T) {
	_, err := LoadManifest("/nonexistent/.parts.yaml")
	if err == nil {
		t.Fatal("Expected error for missing manifest, got nil")
	}
}

func TestManifest_FilterTargets(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, ".parts.yaml")

	yaml := `targets:
  ssh:
    target: /tmp/ssh
    partials: ./ssh/
  vim:
    target: /tmp/vim
    partials: ./vim/
    mode: own
  git:
    target: /tmp/git
    partials: ./git/
`
	if err := os.WriteFile(manifestPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("Failed to write manifest: %v", err)
	}

	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("Failed to load manifest: %v", err)
	}

	// Filter to specific targets
	names, err := manifest.FilterTargets([]string{"ssh", "git"})
	if err != nil {
		t.Fatalf("FilterTargets failed: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("Expected 2 filtered targets, got %d", len(names))
	}

	// Empty filter returns all
	names, err = manifest.FilterTargets(nil)
	if err != nil {
		t.Fatalf("FilterTargets failed: %v", err)
	}
	if len(names) != 3 {
		t.Errorf("Expected 3 targets (all), got %d", len(names))
	}

	// Unknown target name
	_, err = manifest.FilterTargets([]string{"nonexistent"})
	if err == nil {
		t.Fatal("Expected error for unknown target")
	}
}

// containsString is a test helper
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /private/var/www/sandbox/parts && go test ./src/ -run TestLoadManifest -v`
Expected: compilation failure — `LoadManifest` undefined

**Step 3: Write minimal implementation**

Create `src/manifest.go`:

```go
package src

import (
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

// TargetConfig represents a single target in the manifest
type TargetConfig struct {
	Target   string `yaml:"target"`
	Partials string `yaml:"partials"`
	Comment  string `yaml:"comment"`
	Mode     string `yaml:"mode"`
	Backup   *bool  `yaml:"backup"`
}

// ManifestDefaults represents the defaults section of the manifest
type ManifestDefaults struct {
	Comment string `yaml:"comment"`
	Mode    string `yaml:"mode"`
	Backup  bool   `yaml:"backup"`
}

// Manifest represents a parsed .parts.yaml file
type Manifest struct {
	Defaults ManifestDefaults        `yaml:"defaults"`
	Targets  map[string]TargetConfig `yaml:"targets"`
}

// LoadManifest reads and validates a .parts.yaml file
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest '%s': %w", path, err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest '%s': %w", path, err)
	}

	if err := manifest.validate(); err != nil {
		return nil, err
	}

	return &manifest, nil
}

// validate checks the manifest for required fields and valid values
func (m *Manifest) validate() error {
	if len(m.Targets) == 0 {
		return fmt.Errorf("no targets defined in manifest")
	}

	validModes := map[string]bool{"merge": true, "own": true, "": true}

	for name, target := range m.Targets {
		if target.Target == "" {
			return fmt.Errorf("target '%s': missing 'target' path", name)
		}
		if target.Partials == "" {
			return fmt.Errorf("target '%s': missing 'partials' path", name)
		}
		if !validModes[target.Mode] {
			return fmt.Errorf("target '%s': invalid mode '%s' (must be 'merge' or 'own')", name, target.Mode)
		}
	}

	return nil
}

// ResolvedTarget returns a TargetConfig with defaults applied
func (m *Manifest) ResolvedTarget(name string) TargetConfig {
	target := m.Targets[name]

	// Apply defaults for empty fields
	if target.Comment == "" {
		if m.Defaults.Comment != "" {
			target.Comment = m.Defaults.Comment
		} else {
			target.Comment = "auto"
		}
	}

	if target.Mode == "" {
		if m.Defaults.Mode != "" {
			target.Mode = m.Defaults.Mode
		} else {
			target.Mode = "merge"
		}
	}

	if target.Backup == nil {
		backup := m.Defaults.Backup
		target.Backup = &backup
	}

	return target
}

// FilterTargets returns sorted target names, filtered by the given names.
// If names is nil or empty, returns all target names sorted.
// Returns an error if any requested name doesn't exist.
func (m *Manifest) FilterTargets(names []string) ([]string, error) {
	if len(names) == 0 {
		all := make([]string, 0, len(m.Targets))
		for name := range m.Targets {
			all = append(all, name)
		}
		sort.Strings(all)
		return all, nil
	}

	for _, name := range names {
		if _, exists := m.Targets[name]; !exists {
			available := make([]string, 0, len(m.Targets))
			for k := range m.Targets {
				available = append(available, k)
			}
			sort.Strings(available)
			return nil, fmt.Errorf("unknown target '%s' (available: %v)", name, available)
		}
	}

	return names, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `cd /private/var/www/sandbox/parts && go test ./src/ -run TestLoadManifest -v && go test ./src/ -run TestManifest_FilterTargets -v`
Expected: all PASS

**Step 5: Run full test suite to verify no regressions**

Run: `cd /private/var/www/sandbox/parts && go test ./...`
Expected: all PASS

**Step 6: Commit**

```bash
git add src/manifest.go src/manifest_test.go
git commit -m "feat: add manifest parsing and validation for .parts.yaml"
```

---

## Task 3: Own mode engine

**Files:**
- Create: `src/own.go`
- Create: `src/own_test.go`

**Step 1: Write failing tests for own mode**

Create `src/own_test.go`:

```go
package src

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPartialsOwnCommand_Basic(t *testing.T) {
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed to create partials dir: %v", err)
	}
	targetFile := filepath.Join(dir, "target")

	if err := os.WriteFile(filepath.Join(partialsDir, "01-header"), []byte("# Header\n"), 0644); err != nil {
		t.Fatalf("Failed to create partial: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "02-body"), []byte("body content\n"), 0644); err != nil {
		t.Fatalf("Failed to create partial: %v", err)
	}

	cmd := NewPartialsOwnCommand(targetFile, partialsDir, "")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	result, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target: %v", err)
	}

	expected := "# Header\nbody content\n"
	if string(result) != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, string(result))
	}
}

func TestPartialsOwnCommand_WithSourceComments(t *testing.T) {
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed to create partials dir: %v", err)
	}
	targetFile := filepath.Join(dir, "target.sh")

	if err := os.WriteFile(filepath.Join(partialsDir, "part1"), []byte("echo hello\n"), 0644); err != nil {
		t.Fatalf("Failed to create partial: %v", err)
	}

	cmd := NewPartialsOwnCommand(targetFile, partialsDir, "#")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	result, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target: %v", err)
	}

	resultStr := string(result)
	if !containsSubstring(resultStr, "# Source:") {
		t.Error("Expected source comment when comment style is provided")
	}
	if !containsSubstring(resultStr, "echo hello") {
		t.Error("Expected partial content")
	}
}

func TestPartialsOwnCommand_DryRun(t *testing.T) {
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed to create partials dir: %v", err)
	}
	targetFile := filepath.Join(dir, "target")

	if err := os.WriteFile(filepath.Join(partialsDir, "part1"), []byte("content\n"), 0644); err != nil {
		t.Fatalf("Failed to create partial: %v", err)
	}

	cmd := NewPartialsOwnCommand(targetFile, partialsDir, "")
	cmd.SetDryRun(true)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// File should not exist after dry run
	if _, err := os.Stat(targetFile); err == nil {
		t.Error("Target file should not exist after dry run")
	}
}

func TestPartialsOwnCommand_CreatesTargetDir(t *testing.T) {
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed to create partials dir: %v", err)
	}
	// Target in a nested dir that doesn't exist yet
	targetFile := filepath.Join(dir, "nested", "deep", "target")

	if err := os.WriteFile(filepath.Join(partialsDir, "part1"), []byte("content\n"), 0644); err != nil {
		t.Fatalf("Failed to create partial: %v", err)
	}

	cmd := NewPartialsOwnCommand(targetFile, partialsDir, "")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	result, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target: %v", err)
	}
	if string(result) != "content\n" {
		t.Errorf("Unexpected content: %q", string(result))
	}
}

func TestPartialsOwnCommand_PreservesPermissions(t *testing.T) {
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed to create partials dir: %v", err)
	}
	targetFile := filepath.Join(dir, "target")

	// Create existing file with specific permissions
	if err := os.WriteFile(targetFile, []byte("old"), 0755); err != nil {
		t.Fatalf("Failed to create target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "part1"), []byte("new\n"), 0644); err != nil {
		t.Fatalf("Failed to create partial: %v", err)
	}

	cmd := NewPartialsOwnCommand(targetFile, partialsDir, "")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	info, err := os.Stat(targetFile)
	if err != nil {
		t.Fatalf("Failed to stat target: %v", err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("Expected permissions 0755, got %o", info.Mode().Perm())
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /private/var/www/sandbox/parts && go test ./src/ -run TestPartialsOwnCommand -v`
Expected: compilation failure — `NewPartialsOwnCommand` undefined

**Step 3: Write implementation**

Create `src/own.go`:

```go
package src

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// PartialsOwnCommand handles writing entire files from partials (no markers)
type PartialsOwnCommand struct {
	targetFile   string
	partialsDir  string
	commentChars string
	dryRun       bool
}

// NewPartialsOwnCommand creates a new own command.
// commentChars is optional — if non-empty, source-path comments are added.
func NewPartialsOwnCommand(targetFile, partialsDir, commentChars string) PartialsOwnCommand {
	return PartialsOwnCommand{
		targetFile:   targetFile,
		partialsDir:  partialsDir,
		commentChars: commentChars,
	}
}

// SetDryRun sets the dry-run mode
func (p *PartialsOwnCommand) SetDryRun(dryRun bool) {
	p.dryRun = dryRun
}

// Run executes the own command
func (p PartialsOwnCommand) Run() error {
	// Get original file permissions if file exists
	var originalMode fs.FileMode = 0644
	if info, err := os.Stat(p.targetFile); err == nil {
		originalMode = info.Mode()
	}

	files, err := os.ReadDir(p.partialsDir)
	if err != nil {
		return fmt.Errorf("failed to read partials directory '%s': %w", p.partialsDir, err)
	}

	var output strings.Builder

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		partialPath := filepath.Join(p.partialsDir, file.Name())
		content, readErr := os.ReadFile(partialPath)
		if readErr != nil {
			return fmt.Errorf("failed to read partial file '%s': %w", partialPath, readErr)
		}

		// Add source comment if comment style is provided
		if p.commentChars != "" {
			style := ResolveCommentStyle(p.commentChars, p.targetFile)
			if style.End != "" {
				output.WriteString(fmt.Sprintf("%s Source: %s %s\n", style.Start, partialPath, style.End))
			} else {
				output.WriteString(fmt.Sprintf("%s Source: %s\n", style.Start, partialPath))
			}
		}

		output.Write(content)
		// Ensure each partial ends with a newline
		if len(content) > 0 && content[len(content)-1] != '\n' {
			output.WriteByte('\n')
		}
	}

	if p.dryRun {
		fmt.Printf("DRY RUN: Would write to '%s' (own mode)\n", p.targetFile)
		fmt.Printf("Content preview:\n")
		fmt.Printf("--- BEGIN FILE CONTENT ---\n")
		fmt.Print(output.String())
		fmt.Printf("--- END FILE CONTENT ---\n")
		fmt.Printf("Total length: %d characters\n", output.Len())
		return nil
	}

	// Create target directory if it doesn't exist
	targetDir := filepath.Dir(p.targetFile)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory '%s': %w", targetDir, err)
	}

	if err := os.WriteFile(p.targetFile, []byte(output.String()), originalMode); err != nil {
		return fmt.Errorf("failed to write target file '%s': %w", p.targetFile, err)
	}

	// Count partial files for success message
	partialCount := 0
	for _, f := range files {
		if !f.IsDir() {
			partialCount++
		}
	}
	fmt.Printf("Wrote %d partial(s) to '%s' (own mode)\n", partialCount, p.targetFile)

	return nil
}
```

**Step 4: Run tests**

Run: `cd /private/var/www/sandbox/parts && go test ./src/ -run TestPartialsOwnCommand -v`
Expected: all PASS

**Step 5: Run full suite**

Run: `cd /private/var/www/sandbox/parts && go test ./...`
Expected: all PASS

**Step 6: Commit**

```bash
git add src/own.go src/own_test.go
git commit -m "feat: add own mode engine for whole-file management"
```

---

## Task 4: Apply subcommand (cobra wiring + orchestration)

**Files:**
- Create: `cmd/apply.go`
- Create: `cmd/apply_test.go`
- Modify: `cmd/root.go` (register subcommand)

**Step 1: Write failing test for apply subcommand**

Create `cmd/apply_test.go`:

```go
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestApplyCommand_MergeMode(t *testing.T) {
	dir := t.TempDir()

	// Create partials
	partialsDir := filepath.Join(dir, "ssh")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed to create partials dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "work"), []byte("Host work\n    User admin\n"), 0644); err != nil {
		t.Fatalf("Failed to create partial: %v", err)
	}

	// Create target file
	targetFile := filepath.Join(dir, "ssh-config")
	if err := os.WriteFile(targetFile, []byte("# My SSH config\n"), 0644); err != nil {
		t.Fatalf("Failed to create target: %v", err)
	}

	// Create manifest
	manifest := `targets:
  ssh:
    target: ` + targetFile + `
    partials: ` + partialsDir + `
    comment: "#"
    mode: merge
`
	manifestPath := filepath.Join(dir, ".parts.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	// Run apply
	cmd := newApplyCmd()
	cmd.SetArgs([]string{})
	// Override manifest path for testing
	applyManifestPath = manifestPath
	defer func() { applyManifestPath = "" }()

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Verify
	result, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target: %v", err)
	}
	resultStr := string(result)
	if !strings.Contains(resultStr, "# My SSH config") {
		t.Error("Original content not preserved")
	}
	if !strings.Contains(resultStr, "Host work") {
		t.Error("Partial content not merged")
	}
	if !strings.Contains(resultStr, "# PARTIALS>>>>>") {
		t.Error("Markers not present")
	}
}

func TestApplyCommand_OwnMode(t *testing.T) {
	dir := t.TempDir()

	partialsDir := filepath.Join(dir, "vim")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed to create partials dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "vimrc"), []byte("set number\nset tabstop=4\n"), 0644); err != nil {
		t.Fatalf("Failed to create partial: %v", err)
	}

	targetFile := filepath.Join(dir, "vimrc")
	manifest := `targets:
  vim:
    target: ` + targetFile + `
    partials: ` + partialsDir + `
    mode: own
`
	manifestPath := filepath.Join(dir, ".parts.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		t.Fatalf("Failed to create manifest: %v", err)
	}

	cmd := newApplyCmd()
	cmd.SetArgs([]string{})
	applyManifestPath = manifestPath
	defer func() { applyManifestPath = "" }()

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	result, err := os.ReadFile(targetFile)
	if err != nil {
		t.Fatalf("Failed to read target: %v", err)
	}
	resultStr := string(result)
	if !strings.Contains(resultStr, "set number") {
		t.Error("Partial content not written")
	}
	// Own mode should NOT have markers
	if strings.Contains(resultStr, "PARTIALS>>>>>") {
		t.Error("Own mode should not have PARTIALS markers")
	}
}

func TestApplyCommand_SelectiveTarget(t *testing.T) {
	dir := t.TempDir()

	// Create two targets
	sshPartials := filepath.Join(dir, "ssh")
	vimPartials := filepath.Join(dir, "vim")
	if err := os.MkdirAll(sshPartials, 0755); err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if err := os.MkdirAll(vimPartials, 0755); err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sshPartials, "work"), []byte("Host work\n"), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(vimPartials, "rc"), []byte("set number\n"), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	sshTarget := filepath.Join(dir, "ssh-config")
	vimTarget := filepath.Join(dir, "vimrc")
	if err := os.WriteFile(sshTarget, []byte(""), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	manifest := `targets:
  ssh:
    target: ` + sshTarget + `
    partials: ` + sshPartials + `
    comment: "#"
    mode: merge
  vim:
    target: ` + vimTarget + `
    partials: ` + vimPartials + `
    mode: own
`
	manifestPath := filepath.Join(dir, ".parts.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	// Apply only ssh
	cmd := newApplyCmd()
	cmd.SetArgs([]string{"ssh"})
	applyManifestPath = manifestPath
	defer func() { applyManifestPath = "" }()

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// ssh target should be modified
	sshResult, _ := os.ReadFile(sshTarget)
	if !strings.Contains(string(sshResult), "Host work") {
		t.Error("SSH target should have been applied")
	}

	// vim target should NOT exist (we only applied ssh)
	if _, err := os.Stat(vimTarget); err == nil {
		t.Error("Vim target should not have been created")
	}
}

func TestApplyCommand_DryRun(t *testing.T) {
	dir := t.TempDir()

	partialsDir := filepath.Join(dir, "ssh")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "work"), []byte("Host work\n"), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	targetFile := filepath.Join(dir, "ssh-config")
	originalContent := "# Original\n"
	if err := os.WriteFile(targetFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	manifest := `targets:
  ssh:
    target: ` + targetFile + `
    partials: ` + partialsDir + `
    comment: "#"
    mode: merge
`
	manifestPath := filepath.Join(dir, ".parts.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	cmd := newApplyCmd()
	cmd.SetArgs([]string{"--dry-run"})
	applyManifestPath = manifestPath
	defer func() { applyManifestPath = "" }()

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Apply dry-run failed: %v", err)
	}

	// File should be unchanged
	result, _ := os.ReadFile(targetFile)
	if string(result) != originalContent {
		t.Error("File should not be modified in dry-run mode")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /private/var/www/sandbox/parts && go test ./cmd/ -run TestApplyCommand -v`
Expected: compilation failure — `newApplyCmd` undefined

**Step 3: Write the apply subcommand**

Create `cmd/apply.go`:

```go
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/cageis/parts/src"
	"github.com/spf13/cobra"
)

// applyManifestPath allows tests to override the manifest location
var applyManifestPath string

func newApplyCmd() *cobra.Command {
	var applyDryRun bool

	cmd := &cobra.Command{
		Use:   "apply [target-name...]",
		Short: "Apply manifest targets — merge partials into target files",
		Long: `Reads .parts.yaml from the current directory and applies each target.

For 'merge' mode targets, partials are merged into the target file between
PARTIALS markers (existing file content outside the markers is preserved).

For 'own' mode targets, the target file is entirely written from the
concatenated partials (the file is fully managed by Parts).

If target names are specified, only those targets are applied.
If no target names are specified, all targets are applied.`,
		Example: `  parts apply            # Apply all targets
  parts apply ssh        # Apply only the 'ssh' target
  parts apply --dry-run  # Preview changes without modifying files`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestPath := applyManifestPath
			if manifestPath == "" {
				manifestPath = ".parts.yaml"
			}

			absManifest, err := filepath.Abs(manifestPath)
			if err != nil {
				return fmt.Errorf("failed to resolve manifest path: %w", err)
			}

			manifest, err := src.LoadManifest(absManifest)
			if err != nil {
				return err
			}

			names, err := manifest.FilterTargets(args)
			if err != nil {
				return err
			}

			var errors []error
			for _, name := range names {
				target := manifest.ResolvedTarget(name)

				expandedTarget, expandErr := src.ExpandTildePrefix(target.Target)
				if expandErr != nil {
					errors = append(errors, fmt.Errorf("target '%s': %w", name, expandErr))
					continue
				}

				expandedPartials, expandErr := src.ExpandTildePrefix(target.Partials)
				if expandErr != nil {
					errors = append(errors, fmt.Errorf("target '%s': %w", name, expandErr))
					continue
				}

				switch target.Mode {
				case "merge":
					buildCmd, buildErr := src.NewPartialsBuildCommand(expandedTarget, expandedPartials, target.Comment)
					if buildErr != nil {
						errors = append(errors, fmt.Errorf("target '%s': %w", name, buildErr))
						continue
					}
					buildCmd.SetDryRun(applyDryRun)
					if runErr := buildCmd.Run(); runErr != nil {
						errors = append(errors, fmt.Errorf("target '%s': %w", name, runErr))
					}

				case "own":
					ownCmd := src.NewPartialsOwnCommand(expandedTarget, expandedPartials, target.Comment)
					ownCmd.SetDryRun(applyDryRun)
					if runErr := ownCmd.Run(); runErr != nil {
						errors = append(errors, fmt.Errorf("target '%s': %w", name, runErr))
					}
				}
			}

			if len(errors) > 0 {
				for _, e := range errors {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", e)
				}
				return fmt.Errorf("%d target(s) failed", len(errors))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&applyDryRun, "dry-run", "n", false, "preview changes without modifying files")
	return cmd
}
```

**Step 4: Register the subcommand in root.go**

Add to `cmd/root.go` inside the `Execute()` function, before the `rootCmd.Execute()` call:

```go
rootCmd.AddCommand(newApplyCmd())
```

**Step 5: Run tests**

Run: `cd /private/var/www/sandbox/parts && go test ./cmd/ -run TestApplyCommand -v`
Expected: all PASS

**Step 6: Run full suite to verify no regressions (especially existing CLI tests)**

Run: `cd /private/var/www/sandbox/parts && go test ./...`
Expected: all PASS

**Step 7: Commit**

```bash
git add cmd/apply.go cmd/apply_test.go cmd/root.go
git commit -m "feat: add 'apply' subcommand for manifest-driven deployment"
```

---

## Task 5: Manifest-driven remove subcommand

**Files:**
- Create: `cmd/remove_manifest.go`
- Create: `cmd/remove_manifest_test.go`
- Modify: `cmd/root.go` (register subcommand)

**Step 1: Write failing tests**

Create `cmd/remove_manifest_test.go`:

```go
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestManifestRemoveCommand_MergeMode(t *testing.T) {
	dir := t.TempDir()

	partialsDir := filepath.Join(dir, "ssh")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "work"), []byte("Host work\n"), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	targetFile := filepath.Join(dir, "ssh-config")
	if err := os.WriteFile(targetFile, []byte("# My config\n"), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	manifest := `targets:
  ssh:
    target: ` + targetFile + `
    partials: ` + partialsDir + `
    comment: "#"
    mode: merge
`
	manifestPath := filepath.Join(dir, ".parts.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	// First apply
	applyCmd := newApplyCmd()
	applyCmd.SetArgs([]string{})
	applyManifestPath = manifestPath
	if err := applyCmd.Execute(); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Verify markers exist
	content, _ := os.ReadFile(targetFile)
	if !strings.Contains(string(content), "PARTIALS>>>>>") {
		t.Fatal("Markers should exist after apply")
	}

	// Now remove
	rmCmd := newManifestRemoveCmd()
	rmCmd.SetArgs([]string{})
	manifestRemovePath = manifestPath
	defer func() { manifestRemovePath = ""; applyManifestPath = "" }()

	if err := rmCmd.Execute(); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Verify markers are gone but original content preserved
	result, _ := os.ReadFile(targetFile)
	resultStr := string(result)
	if strings.Contains(resultStr, "PARTIALS>>>>>") {
		t.Error("Markers should be removed")
	}
	if !strings.Contains(resultStr, "# My config") {
		t.Error("Original content should be preserved")
	}
}

func TestManifestRemoveCommand_OwnMode(t *testing.T) {
	dir := t.TempDir()

	partialsDir := filepath.Join(dir, "vim")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "rc"), []byte("set number\n"), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	targetFile := filepath.Join(dir, "vimrc")
	manifest := `targets:
  vim:
    target: ` + targetFile + `
    partials: ` + partialsDir + `
    mode: own
`
	manifestPath := filepath.Join(dir, ".parts.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	// Apply first
	applyCmd := newApplyCmd()
	applyCmd.SetArgs([]string{})
	applyManifestPath = manifestPath
	if err := applyCmd.Execute(); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(targetFile); err != nil {
		t.Fatal("Target should exist after apply")
	}

	// Remove
	rmCmd := newManifestRemoveCmd()
	rmCmd.SetArgs([]string{})
	manifestRemovePath = manifestPath
	defer func() { manifestRemovePath = ""; applyManifestPath = "" }()

	if err := rmCmd.Execute(); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// File should be deleted for own mode
	if _, err := os.Stat(targetFile); err == nil {
		t.Error("Target file should be deleted in own mode remove")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /private/var/www/sandbox/parts && go test ./cmd/ -run TestManifestRemoveCommand -v`
Expected: compilation failure

**Step 3: Write implementation**

Create `cmd/remove_manifest.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cageis/parts/src"
	"github.com/spf13/cobra"
)

var manifestRemovePath string

func newManifestRemoveCmd() *cobra.Command {
	var removeDryRun bool

	cmd := &cobra.Command{
		Use:   "remove [target-name...]",
		Short: "Remove managed sections from manifest targets",
		Long: `Reads .parts.yaml and removes the managed content from each target.

For 'merge' mode targets, the PARTIALS markers and their content are removed,
preserving any user content outside the markers.

For 'own' mode targets, the target file is deleted entirely.`,
		Example: `  parts remove           # Remove all targets
  parts remove ssh       # Remove only the 'ssh' target
  parts remove --dry-run # Preview what would be removed`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestPath := manifestRemovePath
			if manifestPath == "" {
				manifestPath = ".parts.yaml"
			}

			absManifest, err := filepath.Abs(manifestPath)
			if err != nil {
				return fmt.Errorf("failed to resolve manifest path: %w", err)
			}

			manifest, err := src.LoadManifest(absManifest)
			if err != nil {
				return err
			}

			names, err := manifest.FilterTargets(args)
			if err != nil {
				return err
			}

			var errors []error
			for _, name := range names {
				target := manifest.ResolvedTarget(name)

				expandedTarget, expandErr := src.ExpandTildePrefix(target.Target)
				if expandErr != nil {
					errors = append(errors, fmt.Errorf("target '%s': %w", name, expandErr))
					continue
				}

				switch target.Mode {
				case "merge":
					rmCmd, rmErr := src.NewPartialsRemoveCommand(expandedTarget, target.Comment)
					if rmErr != nil {
						errors = append(errors, fmt.Errorf("target '%s': %w", name, rmErr))
						continue
					}
					rmCmd.SetDryRun(removeDryRun)
					if runErr := rmCmd.Run(); runErr != nil {
						errors = append(errors, fmt.Errorf("target '%s': %w", name, runErr))
					}

				case "own":
					if removeDryRun {
						fmt.Printf("DRY RUN: Would delete '%s' (own mode)\n", expandedTarget)
					} else {
						if err := os.Remove(expandedTarget); err != nil {
							if !os.IsNotExist(err) {
								errors = append(errors, fmt.Errorf("target '%s': failed to delete '%s': %w", name, expandedTarget, err))
							}
						} else {
							fmt.Printf("Deleted '%s' (own mode)\n", expandedTarget)
						}
					}
				}
			}

			if len(errors) > 0 {
				for _, e := range errors {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", e)
				}
				return fmt.Errorf("%d target(s) failed", len(errors))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&removeDryRun, "dry-run", "n", false, "preview changes without modifying files")
	return cmd
}
```

**Step 4: Register in root.go**

Add alongside the apply registration:
```go
rootCmd.AddCommand(newManifestRemoveCmd())
```

Note: This creates an ambiguity with the `--remove` flag on the root command. The root command uses `--remove` as a flag for legacy mode. The new `remove` subcommand is for manifest mode. Cobra handles this cleanly — `parts remove` invokes the subcommand, `parts --remove file "#"` invokes the root command's flag. No conflict.

**Step 5: Run tests**

Run: `cd /private/var/www/sandbox/parts && go test ./cmd/ -run TestManifestRemoveCommand -v`
Expected: all PASS

**Step 6: Full suite**

Run: `cd /private/var/www/sandbox/parts && go test ./...`
Expected: all PASS

**Step 7: Commit**

```bash
git add cmd/remove_manifest.go cmd/remove_manifest_test.go cmd/root.go
git commit -m "feat: add manifest-driven 'remove' subcommand"
```

---

## Task 6: Init subcommand

**Files:**
- Create: `cmd/init.go`
- Create: `cmd/init_test.go`
- Modify: `cmd/root.go` (register subcommand)

**Step 1: Write failing test**

Create `cmd/init_test.go`:

```go
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCommand_CreatesManifest(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	cmd := newInitCmd()
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	manifestPath := filepath.Join(dir, ".parts.yaml")
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("Manifest not created: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "targets:") {
		t.Error("Manifest should contain 'targets:' section")
	}
	if !strings.Contains(contentStr, "mode:") {
		t.Error("Manifest should contain mode examples")
	}
}

func TestInitCommand_DoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(dir)

	manifestPath := filepath.Join(dir, ".parts.yaml")
	if err := os.WriteFile(manifestPath, []byte("existing content"), 0644); err != nil {
		t.Fatalf("Failed to create existing manifest: %v", err)
	}

	cmd := newInitCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("Expected error when manifest already exists")
	}

	// Content should be unchanged
	content, _ := os.ReadFile(manifestPath)
	if string(content) != "existing content" {
		t.Error("Existing manifest should not be overwritten")
	}
}
```

**Step 2: Run tests to verify failure**

Run: `cd /private/var/www/sandbox/parts && go test ./cmd/ -run TestInitCommand -v`
Expected: compilation failure

**Step 3: Write implementation**

Create `cmd/init.go`:

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
	return &cobra.Command{
		Use:   "init",
		Short: "Generate a skeleton .parts.yaml manifest",
		Long:  `Creates a .parts.yaml file in the current directory with commented examples.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat(".parts.yaml"); err == nil {
				return fmt.Errorf(".parts.yaml already exists in this directory")
			}

			if err := os.WriteFile(".parts.yaml", []byte(manifestTemplate), 0644); err != nil {
				return fmt.Errorf("failed to create .parts.yaml: %w", err)
			}

			fmt.Println("Created .parts.yaml — edit it to define your targets")
			return nil
		},
	}
}
```

**Step 4: Register in root.go**

```go
rootCmd.AddCommand(newInitCmd())
```

**Step 5: Run tests**

Run: `cd /private/var/www/sandbox/parts && go test ./cmd/ -run TestInitCommand -v`
Expected: all PASS

**Step 6: Full suite**

Run: `cd /private/var/www/sandbox/parts && go test ./...`
Expected: all PASS

**Step 7: Commit**

```bash
git add cmd/init.go cmd/init_test.go cmd/root.go
git commit -m "feat: add 'init' subcommand to generate skeleton manifest"
```

---

## Task 7: Sync subcommand

**Files:**
- Create: `src/sync.go`
- Create: `src/sync_test.go`
- Create: `cmd/sync.go`
- Create: `cmd/sync_test.go`
- Modify: `cmd/root.go` (register subcommand)

**Step 1: Write failing tests for sync engine**

Create `src/sync_test.go`:

```go
package src

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractPartialSections_MergeMode(t *testing.T) {
	// Simulate a merged file with source comments
	content := `# My SSH config
Host personal
    User me

# ============================
# PARTIALS>>>>>
# ============================
# Source: /tmp/partials/work.conf
Host work
    User admin
    Port 22
# Source: /tmp/partials/staging.conf
Host staging
    User deploy
# ============================
# PARTIALS<<<<<
# ============================
`

	sections, err := ExtractPartialSections(content, "#")
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	if sections["/tmp/partials/work.conf"] != "Host work\n    User admin\n    Port 22\n" {
		t.Errorf("Unexpected work section content: %q", sections["/tmp/partials/work.conf"])
	}
	if sections["/tmp/partials/staging.conf"] != "Host staging\n    User deploy\n" {
		t.Errorf("Unexpected staging section content: %q", sections["/tmp/partials/staging.conf"])
	}
}

func TestExtractPartialSections_OwnMode(t *testing.T) {
	content := `# Source: /tmp/partials/header
#!/bin/bash
set -e
# Source: /tmp/partials/body
echo "hello"
`
	sections, err := ExtractPartialSections(content, "#")
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(sections) != 2 {
		t.Fatalf("Expected 2 sections, got %d", len(sections))
	}

	if sections["/tmp/partials/header"] != "#!/bin/bash\nset -e\n" {
		t.Errorf("Unexpected header content: %q", sections["/tmp/partials/header"])
	}
}

func TestExtractPartialSections_NoSourceComments(t *testing.T) {
	content := "Host work\n    User admin\n"
	sections, err := ExtractPartialSections(content, "#")
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if len(sections) != 0 {
		t.Errorf("Expected 0 sections when no source comments, got %d", len(sections))
	}
}

func TestSyncTarget_MergeMode(t *testing.T) {
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	// Create initial partials
	if err := os.WriteFile(filepath.Join(partialsDir, "work"), []byte("Host work\n    User admin\n"), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	// Create target and apply
	targetFile := filepath.Join(dir, "config")
	if err := os.WriteFile(targetFile, []byte("# My config\n"), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	buildCmd, _ := NewPartialsBuildCommand(targetFile, partialsDir, "#")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Now simulate user editing the target file (change User admin -> User root)
	content, _ := os.ReadFile(targetFile)
	modified := strings.Replace(string(content), "User admin", "User root", 1)
	if err := os.WriteFile(targetFile, []byte(modified), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	// Sync back
	result, err := SyncTarget(targetFile, partialsDir, "#", "merge", false)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if result.UpdatedFiles != 1 {
		t.Errorf("Expected 1 updated file, got %d", result.UpdatedFiles)
	}

	// Verify partial was updated
	partialContent, _ := os.ReadFile(filepath.Join(partialsDir, "work"))
	if !strings.Contains(string(partialContent), "User root") {
		t.Error("Partial should contain updated content")
	}
}

func TestSyncTarget_DryRun(t *testing.T) {
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(partialsDir, "work"), []byte("Host work\n    User admin\n"), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	targetFile := filepath.Join(dir, "config")
	if err := os.WriteFile(targetFile, []byte("# My config\n"), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	buildCmd, _ := NewPartialsBuildCommand(targetFile, partialsDir, "#")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Modify target
	content, _ := os.ReadFile(targetFile)
	modified := strings.Replace(string(content), "User admin", "User root", 1)
	os.WriteFile(targetFile, []byte(modified), 0644)

	// Dry-run sync
	result, err := SyncTarget(targetFile, partialsDir, "#", "merge", true)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if result.UpdatedFiles != 1 {
		t.Errorf("Dry run should still report 1 changed file, got %d", result.UpdatedFiles)
	}

	// Partial should NOT be modified
	partialContent, _ := os.ReadFile(filepath.Join(partialsDir, "work"))
	if strings.Contains(string(partialContent), "User root") {
		t.Error("Partial should not be modified in dry-run mode")
	}
}
```

**Step 2: Run tests to verify failure**

Run: `cd /private/var/www/sandbox/parts && go test ./src/ -run "TestExtractPartialSections|TestSyncTarget" -v`
Expected: compilation failure

**Step 3: Write sync engine**

Create `src/sync.go`:

```go
package src

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SyncResult reports what the sync operation found/did
type SyncResult struct {
	UpdatedFiles  int
	SkippedFiles  int
	ChangedPaths  []string
}

// ExtractPartialSections parses file content and splits it by "# Source: <path>" comments.
// Returns a map of source-path -> content-after-that-comment.
func ExtractPartialSections(content, commentChars string) (map[string]string, error) {
	style := ResolveCommentStyle(commentChars, "")
	prefix := fmt.Sprintf("%s Source: ", style.Start)

	lines := strings.Split(content, "\n")
	sections := make(map[string]string)
	var currentPath string
	var currentContent strings.Builder

	for _, line := range lines {
		trimmed := line
		// Check for source comment
		if strings.HasPrefix(trimmed, prefix) {
			// Save previous section
			if currentPath != "" {
				sections[currentPath] = currentContent.String()
			}
			// Extract path from source comment
			pathPart := strings.TrimPrefix(trimmed, prefix)
			// Remove closing comment chars if present (e.g., " */")
			if style.End != "" {
				pathPart = strings.TrimSuffix(pathPart, " "+style.End)
			}
			currentPath = strings.TrimSpace(pathPart)
			currentContent.Reset()
			continue
		}

		// Skip marker lines (PARTIALS>>>>>, PARTIALS<<<<<, separator)
		if strings.Contains(line, PartialStartMarker) ||
			strings.Contains(line, PartialEndMarker) ||
			strings.Contains(line, MarkerSeparator) {
			continue
		}

		if currentPath != "" {
			currentContent.WriteString(line)
			currentContent.WriteByte('\n')
		}
	}

	// Save last section
	if currentPath != "" {
		sections[currentPath] = currentContent.String()
	}

	return sections, nil
}

// SyncTarget reads the target file, extracts sections by source comment,
// and writes changed content back to the partial files.
func SyncTarget(targetFile, partialsDir, commentChars, mode string, dryRun bool) (*SyncResult, error) {
	content, err := os.ReadFile(targetFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read target file '%s': %w", targetFile, err)
	}

	var sectionContent string
	contentStr := string(content)

	if mode == "merge" {
		// Extract only the content between PARTIALS markers
		style := ResolveCommentStyle(commentChars, targetFile)
		var startFlag, endFlag string
		if style.End != "" {
			startFlag = fmt.Sprintf("%s\n%s %s\n%s", style.Start, style.Start, PartialStartMarker, style.End)
			endFlag = fmt.Sprintf("%s\n%s %s\n%s", style.Start, style.Start, PartialEndMarker, style.End)
		} else {
			startFlag = fmt.Sprintf("%s %s\n%s %s\n%s %s",
				style.Start, MarkerSeparator, style.Start, PartialStartMarker, style.Start, MarkerSeparator)
			endFlag = fmt.Sprintf("%s %s\n%s %s\n%s %s",
				style.Start, MarkerSeparator, style.Start, PartialEndMarker, style.Start, MarkerSeparator)
		}

		startIdx := strings.Index(contentStr, startFlag)
		endIdx := strings.Index(contentStr, endFlag)
		if startIdx == -1 || endIdx == -1 {
			return &SyncResult{}, nil // No managed section found
		}
		sectionContent = contentStr[startIdx:endIdx]
	} else {
		// Own mode: entire file is managed
		sectionContent = contentStr
	}

	sections, err := ExtractPartialSections(sectionContent, commentChars)
	if err != nil {
		return nil, fmt.Errorf("failed to extract sections: %w", err)
	}

	result := &SyncResult{}

	for sourcePath, newContent := range sections {
		// Verify the source path is within the partials directory
		absSource, _ := filepath.Abs(sourcePath)
		absPartials, _ := filepath.Abs(partialsDir)
		if !strings.HasPrefix(absSource, absPartials) {
			result.SkippedFiles++
			continue
		}

		// Read current partial content
		existing, readErr := os.ReadFile(sourcePath)
		if readErr != nil {
			result.SkippedFiles++
			continue
		}

		// Compare
		// Trim trailing newlines for comparison to handle minor formatting diffs
		existingTrimmed := strings.TrimRight(string(existing), "\n")
		newTrimmed := strings.TrimRight(newContent, "\n")

		if existingTrimmed == newTrimmed {
			continue // No change
		}

		result.UpdatedFiles++
		result.ChangedPaths = append(result.ChangedPaths, sourcePath)

		if dryRun {
			fmt.Printf("DRY RUN: Would update '%s'\n", sourcePath)
			continue
		}

		// Write back — preserve trailing newline
		writeContent := strings.TrimRight(newContent, "\n") + "\n"
		if writeErr := os.WriteFile(sourcePath, []byte(writeContent), 0644); writeErr != nil {
			return nil, fmt.Errorf("failed to write partial '%s': %w", sourcePath, writeErr)
		}
		fmt.Printf("Updated '%s'\n", sourcePath)
	}

	return result, nil
}
```

**Step 4: Run sync engine tests**

Run: `cd /private/var/www/sandbox/parts && go test ./src/ -run "TestExtractPartialSections|TestSyncTarget" -v`
Expected: all PASS

**Step 5: Write failing test for sync subcommand**

Create `cmd/sync_test.go`:

```go
package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncCommand_MergeMode(t *testing.T) {
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "ssh")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "work"), []byte("Host work\n    User admin\n"), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	targetFile := filepath.Join(dir, "ssh-config")
	if err := os.WriteFile(targetFile, []byte("# My config\n"), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	manifest := `targets:
  ssh:
    target: ` + targetFile + `
    partials: ` + partialsDir + `
    comment: "#"
    mode: merge
`
	manifestPath := filepath.Join(dir, ".parts.yaml")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0644); err != nil {
		t.Fatalf("Failed: %v", err)
	}

	// Apply first
	applyCmd := newApplyCmd()
	applyCmd.SetArgs([]string{})
	applyManifestPath = manifestPath
	if err := applyCmd.Execute(); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Modify the target file
	content, _ := os.ReadFile(targetFile)
	modified := strings.Replace(string(content), "User admin", "User root", 1)
	os.WriteFile(targetFile, []byte(modified), 0644)

	// Sync
	syncCmd := newSyncCmd()
	syncCmd.SetArgs([]string{})
	syncManifestPath = manifestPath
	defer func() { syncManifestPath = ""; applyManifestPath = "" }()

	if err := syncCmd.Execute(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	// Verify partial was updated
	partialContent, _ := os.ReadFile(filepath.Join(partialsDir, "work"))
	if !strings.Contains(string(partialContent), "User root") {
		t.Error("Partial should be updated with synced content")
	}
}
```

**Step 6: Write sync subcommand**

Create `cmd/sync.go`:

```go
package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/cageis/parts/src"
	"github.com/spf13/cobra"
)

var syncManifestPath string

func newSyncCmd() *cobra.Command {
	var syncDryRun bool

	cmd := &cobra.Command{
		Use:   "sync [target-name...]",
		Short: "Sync changes from target files back into partials",
		Long: `Reads .parts.yaml and detects changes in target files, pulling modified
content back into the partial source files.

Uses the '# Source: <path>' comments to map content back to individual
partial files.`,
		Example: `  parts sync            # Sync all targets
  parts sync ssh        # Sync only the 'ssh' target
  parts sync --dry-run  # Preview what would be synced`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestPath := syncManifestPath
			if manifestPath == "" {
				manifestPath = ".parts.yaml"
			}

			absManifest, err := filepath.Abs(manifestPath)
			if err != nil {
				return fmt.Errorf("failed to resolve manifest path: %w", err)
			}

			manifest, err := src.LoadManifest(absManifest)
			if err != nil {
				return err
			}

			names, err := manifest.FilterTargets(args)
			if err != nil {
				return err
			}

			var errors []error
			totalUpdated := 0

			for _, name := range names {
				target := manifest.ResolvedTarget(name)

				expandedTarget, expandErr := src.ExpandTildePrefix(target.Target)
				if expandErr != nil {
					errors = append(errors, fmt.Errorf("target '%s': %w", name, expandErr))
					continue
				}

				expandedPartials, expandErr := src.ExpandTildePrefix(target.Partials)
				if expandErr != nil {
					errors = append(errors, fmt.Errorf("target '%s': %w", name, expandErr))
					continue
				}

				result, syncErr := src.SyncTarget(expandedTarget, expandedPartials, target.Comment, target.Mode, syncDryRun)
				if syncErr != nil {
					errors = append(errors, fmt.Errorf("target '%s': %w", name, syncErr))
					continue
				}

				totalUpdated += result.UpdatedFiles
			}

			if syncDryRun {
				fmt.Printf("DRY RUN: %d partial file(s) would be updated\n", totalUpdated)
			} else if totalUpdated == 0 {
				fmt.Println("All partials are in sync")
			}

			if len(errors) > 0 {
				for _, e := range errors {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", e)
				}
				return fmt.Errorf("%d target(s) failed", len(errors))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&syncDryRun, "dry-run", "n", false, "preview changes without modifying files")
	return cmd
}
```

**Step 7: Register in root.go**

```go
rootCmd.AddCommand(newSyncCmd())
```

**Step 8: Run all sync tests**

Run: `cd /private/var/www/sandbox/parts && go test ./... -run "Sync|Extract" -v`
Expected: all PASS

**Step 9: Full suite**

Run: `cd /private/var/www/sandbox/parts && go test ./...`
Expected: all PASS

**Step 10: Commit**

```bash
git add src/sync.go src/sync_test.go cmd/sync.go cmd/sync_test.go cmd/root.go
git commit -m "feat: add 'sync' subcommand for reverse-syncing target changes to partials"
```

---

## Task 8: Verify backward compatibility

**Files:** None created/modified — verification only

**Step 1: Build the binary**

Run: `cd /private/var/www/sandbox/parts && go build -o bin/parts ./main.go`
Expected: clean build

**Step 2: Verify old-style CLI still works**

Run manual tests:
```bash
# Create test fixtures
mkdir -p /tmp/parts-test/partials
echo "Host test" > /tmp/parts-test/partials/test.conf
echo "# Original" > /tmp/parts-test/config

# Old-style build
bin/parts /tmp/parts-test/config /tmp/parts-test/partials "#"

# Old-style dry-run
bin/parts --dry-run /tmp/parts-test/config /tmp/parts-test/partials "#"

# Old-style remove
bin/parts --remove /tmp/parts-test/config "#"

# Cleanup
rm -rf /tmp/parts-test
```

Expected: all commands work exactly as before

**Step 3: Verify new subcommands**

```bash
bin/parts --help         # Should show apply, remove, sync, init subcommands
bin/parts apply --help   # Should show apply usage
bin/parts init --help    # Should show init usage
```

**Step 4: Run full test suite one final time**

Run: `cd /private/var/www/sandbox/parts && go test ./... -v`
Expected: all PASS

**Step 5: Commit (if any fixes were needed)**

Only commit if fixes were required. Otherwise this task is pure verification.

---

## Task 9: Update documentation

**Files:**
- Modify: `CLAUDE.md` — update Architecture, Commands, and Usage sections
- Modify: `ROADMAP.md` — mark relevant items as done, add new items

**Step 1: Update CLAUDE.md with new commands and architecture**

Add to the Commands section:
```
### Manifest Mode
- `bin/parts init` — Generate skeleton .parts.yaml
- `bin/parts apply` — Apply all manifest targets
- `bin/parts apply ssh` — Apply specific target
- `bin/parts apply --dry-run` — Preview changes
- `bin/parts remove` — Remove managed content from all targets
- `bin/parts sync` — Sync target changes back to partials
```

Update Architecture section to include new files.

**Step 2: Update ROADMAP.md**

Mark completed items and add new roadmap items for sync improvements.

**Step 3: Commit**

```bash
git add CLAUDE.md ROADMAP.md
git commit -m "docs: update documentation for manifest-driven dotfile management"
```
