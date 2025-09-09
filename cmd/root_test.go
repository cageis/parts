package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommand_Help(t *testing.T) {
	// Test that help flag shows proper usage using the actual root command
	helpOutput := rootCmd.UsageString()

	if !strings.Contains(helpOutput, "aggregate-file") {
		t.Error("Help should contain aggregate-file parameter")
	}
	if !strings.Contains(helpOutput, "partials-directory") {
		t.Error("Help should contain partials-directory parameter")
	}
	if !strings.Contains(helpOutput, "comment-style") {
		t.Error("Help should contain comment-style parameter")
	}
	if !strings.Contains(helpOutput, "dry-run") {
		t.Error("Help should contain dry-run flag")
	}
}

func TestRootCommand_ArgumentValidation(t *testing.T) {
	// Test insufficient arguments using actual root command
	// Create a copy to avoid affecting other tests
	cmd := &cobra.Command{
		Use:  rootCmd.Use,
		Args: rootCmd.Args,
		RunE: rootCmd.RunE,
	}
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "preview changes")

	// Test with no arguments
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error with no arguments")
	}
	if !strings.Contains(err.Error(), "build mode requires exactly 3 arguments") {
		t.Errorf("Expected argument count error, got: %v", err)
	}
}

func TestRunParts_DryRun(t *testing.T) {
	// Given
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed to create partials directory: %v", err)
	}
	aggregateFile := filepath.Join(dir, "agg")

	originalContent := "# Original config\n"
	if err := os.WriteFile(aggregateFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create aggregate file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "partial1"), []byte("Host test"), 0644); err != nil {
		t.Fatalf("Failed to create partial file: %v", err)
	}

	// When - run with dry-run flag using actual root command
	// Reset dryRun to false first
	originalDryRun := dryRun
	dryRun = false
	defer func() { dryRun = originalDryRun }()

	// Create a copy to avoid affecting other tests
	cmd := &cobra.Command{
		Use:  rootCmd.Use,
		Args: rootCmd.Args,
		RunE: rootCmd.RunE,
	}
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "preview changes")

	// Set the dry-run flag
	cmd.SetArgs([]string{"--dry-run", aggregateFile, partialsDir, "#"})

	// Capture output
	var output bytes.Buffer
	cmd.SetOut(&output)

	err := cmd.Execute()

	// Then
	if err != nil {
		t.Fatalf("Command should not fail: %v", err)
	}

	// File should be unchanged
	actualContent, err := os.ReadFile(aggregateFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(actualContent) != originalContent {
		t.Errorf("File was modified in dry-run mode!")
	}
}

func TestRunParts_Normal(t *testing.T) {
	// Given
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed to create partials directory: %v", err)
	}
	aggregateFile := filepath.Join(dir, "agg")

	originalContent := "# Original config\n"
	if err := os.WriteFile(aggregateFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create aggregate file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "partial1"), []byte("Host test"), 0644); err != nil {
		t.Fatalf("Failed to create partial file: %v", err)
	}

	// When - run without dry-run flag using actual root command
	// Reset dryRun to false first
	originalDryRun := dryRun
	dryRun = false
	defer func() { dryRun = originalDryRun }()

	// Create a copy to avoid affecting other tests
	cmd := &cobra.Command{
		Use:  rootCmd.Use,
		Args: rootCmd.Args,
		RunE: rootCmd.RunE,
	}
	cmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "preview changes")

	cmd.SetArgs([]string{aggregateFile, partialsDir, "#"})

	err := cmd.Execute()

	// Then
	if err != nil {
		t.Fatalf("Command should not fail: %v", err)
	}

	// File should be modified
	actualContent, err := os.ReadFile(aggregateFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(actualContent) == originalContent {
		t.Error("File should have been modified in normal mode")
	}

	if !strings.Contains(string(actualContent), "# PARTIALS>>>>>") {
		t.Error("File should contain partials markers")
	}
}
