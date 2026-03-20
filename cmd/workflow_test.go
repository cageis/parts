package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

	// Override manifest path for all commands
	manifestPath := filepath.Join(dir, ".parts.yaml")
	initManifestPath = manifestPath
	applyManifestPath = manifestPath
	manifestRemovePath = manifestPath
	defer func() {
		initManifestPath = ""
		applyManifestPath = ""
		manifestRemovePath = ""
	}()

	// Step 1: init --from
	initCmd := newInitCmd()
	initCmd.SetArgs([]string{"--from", targetFile, "--from", partialsDir, "--from", "#", "--name", "ssh"})
	if err := initCmd.Execute(); err != nil {
		t.Fatalf("init --from failed: %v", err)
	}

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
