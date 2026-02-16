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

func TestLoadManifest_TildePaths(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, ".parts.yaml")

	yaml := `targets:
  ssh:
    target: ~/.ssh/config
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
