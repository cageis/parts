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

func TestSyncCommand_NoChanges(t *testing.T) {
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

	// Apply
	applyCmd := newApplyCmd()
	applyCmd.SetArgs([]string{})
	applyManifestPath = manifestPath
	if err := applyCmd.Execute(); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Sync without changes — should succeed with "in sync" message
	syncCmd := newSyncCmd()
	syncCmd.SetArgs([]string{})
	syncManifestPath = manifestPath
	defer func() { syncManifestPath = ""; applyManifestPath = "" }()

	if err := syncCmd.Execute(); err != nil {
		t.Fatalf("Sync failed: %v", err)
	}
}

func TestSyncCommand_DryRun(t *testing.T) {
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

	// Apply
	applyCmd := newApplyCmd()
	applyCmd.SetArgs([]string{})
	applyManifestPath = manifestPath
	if err := applyCmd.Execute(); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Modify target
	content, _ := os.ReadFile(targetFile)
	modified := strings.Replace(string(content), "User admin", "User root", 1)
	os.WriteFile(targetFile, []byte(modified), 0644)

	// Dry-run sync
	syncCmd := newSyncCmd()
	syncCmd.SetArgs([]string{"--dry-run"})
	syncManifestPath = manifestPath
	defer func() { syncManifestPath = ""; applyManifestPath = "" }()

	if err := syncCmd.Execute(); err != nil {
		t.Fatalf("Sync dry-run failed: %v", err)
	}

	// Partial should NOT be modified
	partialContent, _ := os.ReadFile(filepath.Join(partialsDir, "work"))
	if strings.Contains(string(partialContent), "User root") {
		t.Error("Partial should not be modified in dry-run mode")
	}
}
