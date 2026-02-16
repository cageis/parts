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

func TestManifestRemoveCommand_SelectiveTarget(t *testing.T) {
	dir := t.TempDir()

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
	if err := os.WriteFile(sshTarget, []byte("# SSH\n"), 0644); err != nil {
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

	// Apply all first
	applyCmd := newApplyCmd()
	applyCmd.SetArgs([]string{})
	applyManifestPath = manifestPath
	if err := applyCmd.Execute(); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Remove only ssh
	rmCmd := newManifestRemoveCmd()
	rmCmd.SetArgs([]string{"ssh"})
	manifestRemovePath = manifestPath
	defer func() { manifestRemovePath = ""; applyManifestPath = "" }()

	if err := rmCmd.Execute(); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// SSH markers should be gone
	sshContent, _ := os.ReadFile(sshTarget)
	if strings.Contains(string(sshContent), "PARTIALS>>>>>") {
		t.Error("SSH markers should be removed")
	}

	// Vim file should still exist (we only removed ssh)
	if _, err := os.Stat(vimTarget); err != nil {
		t.Error("Vim target should still exist (only ssh was removed)")
	}
}
