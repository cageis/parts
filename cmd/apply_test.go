package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
