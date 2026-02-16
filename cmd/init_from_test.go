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
		{"/etc/hosts", "etc-hosts"},
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
