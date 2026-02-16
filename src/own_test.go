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
