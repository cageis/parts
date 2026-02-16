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

func TestSyncTarget_NoChanges(t *testing.T) {
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

	// Sync without changes
	result, err := SyncTarget(targetFile, partialsDir, "#", "merge", false)
	if err != nil {
		t.Fatalf("Sync failed: %v", err)
	}

	if result.UpdatedFiles != 0 {
		t.Errorf("Expected 0 updated files when nothing changed, got %d", result.UpdatedFiles)
	}
}
