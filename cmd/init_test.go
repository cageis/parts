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
