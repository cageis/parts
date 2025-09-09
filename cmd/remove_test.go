package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRemoveCommand_ArgumentValidation(t *testing.T) {
	// Create a copy to avoid affecting other tests
	cmd := &cobra.Command{
		Use:  rootCmd.Use,
		Args: rootCmd.Args,
		RunE: rootCmd.RunE,
	}
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "preview changes")
	cmd.Flags().BoolVarP(&remove, "remove", "r", false, "remove partials section")

	// Test remove mode with wrong number of arguments
	remove = true
	defer func() { remove = false }()

	// Test with no arguments
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error with no arguments in remove mode")
	}
	if !strings.Contains(err.Error(), "remove mode requires exactly 2 arguments") {
		t.Errorf("Expected remove mode argument error, got: %v", err)
	}
}

func TestRemoveCommand_NoPartialsSection(t *testing.T) {
	// Given - file without partials section
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.conf")
	originalContent := "# Original config\nHost example\n    User test\n"
	
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// When - create remove command
	cmd := &cobra.Command{
		Use:  rootCmd.Use,
		Args: rootCmd.Args,
		RunE: rootCmd.RunE,
	}
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "preview changes")
	cmd.Flags().BoolVarP(&remove, "remove", "r", false, "remove partials section")

	// Set remove mode and dry-run
	originalRemove := remove
	originalDryRun := dryRun
	remove = true
	dryRun = true
	defer func() { 
		remove = originalRemove
		dryRun = originalDryRun 
	}()

	cmd.SetArgs([]string{testFile, "#"})
	err := cmd.Execute()

	// Then - should succeed (no error) with dry-run message
	if err != nil {
		t.Errorf("Dry-run remove should not fail for file without partials: %v", err)
	}

	// File should be unchanged
	actualContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}
	if string(actualContent) != originalContent {
		t.Error("File should be unchanged in dry-run mode")
	}
}

func TestRemoveCommand_WithPartialsSection(t *testing.T) {
	// Given - file with partials section
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test.conf")
	
	contentWithPartials := `# Original config
Host example
    User test
# ============================
# PARTIALS>>>>>
# ============================
Host server1
    User admin
# ============================
# PARTIALS<<<<<
# ============================
`

	if err := os.WriteFile(testFile, []byte(contentWithPartials), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// When - create remove command
	cmd := &cobra.Command{
		Use:  rootCmd.Use,
		Args: rootCmd.Args,
		RunE: rootCmd.RunE,
	}
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "preview changes")
	cmd.Flags().BoolVarP(&remove, "remove", "r", false, "remove partials section")

	// Set remove mode (not dry-run)
	originalRemove := remove
	originalDryRun := dryRun
	remove = true
	dryRun = false
	defer func() { 
		remove = originalRemove
		dryRun = originalDryRun 
	}()

	cmd.SetArgs([]string{testFile, "#"})
	err := cmd.Execute()

	// Then - should succeed
	if err != nil {
		t.Fatalf("Remove command should not fail: %v", err)
	}

	// File should have partials section removed
	actualContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read result file: %v", err)
	}

	actualStr := string(actualContent)
	if strings.Contains(actualStr, "PARTIALS>>>>>") {
		t.Error("Partials section should have been removed")
	}
	if !strings.Contains(actualStr, "# Original config") {
		t.Error("Original content should be preserved")
	}
	if !strings.Contains(actualStr, "Host example") {
		t.Error("Original host config should be preserved")
	}
}

func TestRemoveCommand_DifferentCommentStyles(t *testing.T) {
	tests := []struct {
		name         string
		commentStyle string
		fileContent  string
	}{
		{
			name:         "Hash comments",
			commentStyle: "#",
			fileContent: `# Original
# ============================
# PARTIALS>>>>>
# ============================
content
# ============================
# PARTIALS<<<<<
# ============================
`,
		},
		{
			name:         "Block comments",
			commentStyle: "/*",
			fileContent: `/* Original */
/*
/* PARTIALS>>>>>
*/
content
/*
/* PARTIALS<<<<<
*/
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			dir := t.TempDir()
			testFile := filepath.Join(dir, "test")

			if err := os.WriteFile(testFile, []byte(test.fileContent), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// When - create remove command
			cmd := &cobra.Command{
				Use:  rootCmd.Use,
				Args: rootCmd.Args,
				RunE: rootCmd.RunE,
			}
			cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "preview changes")
			cmd.Flags().BoolVarP(&remove, "remove", "r", false, "remove partials section")

			// Set remove mode with dry-run
			originalRemove := remove
			originalDryRun := dryRun
			remove = true
			dryRun = true
			defer func() { 
				remove = originalRemove
				dryRun = originalDryRun 
			}()

			cmd.SetArgs([]string{testFile, test.commentStyle})
			err := cmd.Execute()

			// Then
			if err != nil {
				t.Fatalf("Remove command should not fail: %v", err)
			}
		})
	}
}