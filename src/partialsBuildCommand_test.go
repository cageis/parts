package src

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPartialsBuildCommand_BasicFunctionality(t *testing.T) {
	// Given
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed to create partials directory: %v", err)
	}
	aggregateFile := filepath.Join(dir, "agg")

	// Create partial files
	if err := os.WriteFile(filepath.Join(partialsDir, "partials1"), []byte("Partial 1"), 0644); err != nil {
		t.Fatalf("Failed to create partial file 1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "partials2"), []byte("Partial 2"), 0644); err != nil {
		t.Fatalf("Failed to create partial file 2: %v", err)
	}
	if err := os.WriteFile(aggregateFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create aggregate file: %v", err)
	}

	// When
	command := NewPartialsBuildCommand(aggregateFile, partialsDir, "#")
	err := command.Run()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Then
	actual, err := os.ReadFile(aggregateFile)
	if err != nil {
		t.Fatalf("Failed to read result file: %v", err)
	}
	expected := "\n# ============================\n# PARTIALS>>>>>\n# ============================\nPartial 1\nPartial 2\n# ============================\n# PARTIALS<<<<<\n# ============================\n"

	if string(actual) != expected {
		t.Errorf("Expected:\n%s\nActual:\n%s", expected, string(actual))
	}
}

func TestPartialsBuildCommand_IdempotentEditing(t *testing.T) {
	// Given - setup with existing content
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed to create partials directory: %v", err)
	}
	aggregateFile := filepath.Join(dir, "agg")

	// Start with some original content and existing partial files
	originalContent := "# Original SSH config\nHost example.com\n    User test\n"
	if err := os.WriteFile(aggregateFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create aggregate file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "partial1"), []byte("Host server1\n    User admin"), 0644); err != nil {
		t.Fatalf("Failed to create partial file 1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "partial2"), []byte("Host server2\n    User root"), 0644); err != nil {
		t.Fatalf("Failed to create partial file 2: %v", err)
	}

	command := NewPartialsBuildCommand(aggregateFile, partialsDir, "#")

	// When - run command multiple times
	if err := command.Run(); err != nil {
		t.Fatalf("First run failed: %v", err)
	}
	firstResult, err := os.ReadFile(aggregateFile)
	if err != nil {
		t.Fatalf("Failed to read first result: %v", err)
	}
	firstLength := len(firstResult)

	if err := command.Run(); err != nil {
		t.Fatalf("Second run failed: %v", err)
	}
	secondResult, err := os.ReadFile(aggregateFile)
	if err != nil {
		t.Fatalf("Failed to read second result: %v", err)
	}
	secondLength := len(secondResult)

	if err := command.Run(); err != nil {
		t.Fatalf("Third run failed: %v", err)
	}
	thirdResult, err := os.ReadFile(aggregateFile)
	if err != nil {
		t.Fatalf("Failed to read third result: %v", err)
	}
	thirdLength := len(thirdResult)

	// Then - all results should be identical (no incremental changes)
	if string(firstResult) != string(secondResult) {
		t.Errorf("First run differs from second run:\nFirst: %q\nSecond: %q", string(firstResult), string(secondResult))
	}
	if string(secondResult) != string(thirdResult) {
		t.Errorf("Second run differs from third run:\nSecond: %q\nThird: %q", string(secondResult), string(thirdResult))
	}

	// File length should remain constant (no whitespace growth)
	if firstLength != secondLength || secondLength != thirdLength {
		t.Errorf("File length growing: %d -> %d -> %d", firstLength, secondLength, thirdLength)
	}

	// Test adding new partial files
	if err := os.WriteFile(filepath.Join(partialsDir, "partial3"), []byte("Host server3\n    User guest"), 0644); err != nil {
		t.Fatalf("Failed to create partial file 3: %v", err)
	}
	if err := command.Run(); err != nil {
		t.Fatalf("Run with new partial failed: %v", err)
	}
	resultWithNewPartial, err := os.ReadFile(aggregateFile)
	if err != nil {
		t.Fatalf("Failed to read result with new partial: %v", err)
	}
	lengthWithNewPartial := len(resultWithNewPartial)

	// Run again to ensure still idempotent after adding new content
	if err := command.Run(); err != nil {
		t.Fatalf("Final run failed: %v", err)
	}
	finalResult, err := os.ReadFile(aggregateFile)
	if err != nil {
		t.Fatalf("Failed to read final result: %v", err)
	}
	finalLength := len(finalResult)

	if string(resultWithNewPartial) != string(finalResult) {
		t.Errorf("Results differ after adding new partial file:\nBefore: %q\nAfter: %q",
			string(resultWithNewPartial), string(finalResult))
	}

	// Length should remain constant after adding new content
	if lengthWithNewPartial != finalLength {
		t.Errorf("File length growing after new partial: %d -> %d", lengthWithNewPartial, finalLength)
	}
}

func TestPartialsBuildCommand_ErrorHandling(t *testing.T) {
	t.Run("NonexistentAggregateFile", func(t *testing.T) {
		command := NewPartialsBuildCommand("/nonexistent/file", "/tmp", "#")
		err := command.Run()
		if err == nil {
			t.Error("Expected error for nonexistent aggregate file, got nil")
		}
		if !strings.Contains(err.Error(), "failed to read aggregate file") {
			t.Errorf("Expected error message about aggregate file, got: %v", err)
		}
	})

	t.Run("NonexistentPartialsDirectory", func(t *testing.T) {
		dir := t.TempDir()
		aggregateFile := filepath.Join(dir, "agg")
		if err := os.WriteFile(aggregateFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create aggregate file: %v", err)
		}

		command := NewPartialsBuildCommand(aggregateFile, "/nonexistent/dir", "#")
		err := command.Run()
		if err == nil {
			t.Error("Expected error for nonexistent partials directory, got nil")
		}
		if !strings.Contains(err.Error(), "failed to read partials directory") {
			t.Errorf("Expected error message about partials directory, got: %v", err)
		}
	})
}

func TestPartialsBuildCommand_PreservesOriginalContent(t *testing.T) {
	// Given
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed to create partials directory: %v", err)
	}
	aggregateFile := filepath.Join(dir, "agg")

	originalContent := "# My existing config\nHost personal\n    User me\n"
	if err := os.WriteFile(aggregateFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create aggregate file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "work"), []byte("Host work\n    User admin"), 0644); err != nil {
		t.Fatalf("Failed to create partial file: %v", err)
	}

	// When
	command := NewPartialsBuildCommand(aggregateFile, partialsDir, "#")
	if err := command.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Then
	result, err := os.ReadFile(aggregateFile)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}

	resultStr := string(result)
	if !strings.Contains(resultStr, "# My existing config") {
		t.Error("Original content not preserved")
	}
	if !strings.Contains(resultStr, "Host personal") {
		t.Error("Original host config not preserved")
	}
	if !strings.Contains(resultStr, "Host work") {
		t.Error("Partial content not included")
	}
	if !strings.Contains(resultStr, "# PARTIALS>>>>>") || !strings.Contains(resultStr, "# PARTIALS<<<<<") {
		t.Error("Partials markers not found")
	}
	if !strings.Contains(resultStr, "# ============================") {
		t.Error("Enhanced comment block markers not found")
	}
}

func TestPartialsBuildCommand_DryRun(t *testing.T) {
	// Given
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed to create partials directory: %v", err)
	}
	aggregateFile := filepath.Join(dir, "agg")

	originalContent := "# Original SSH config\nHost example.com\n    User test\n"
	if err := os.WriteFile(aggregateFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Failed to create aggregate file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "partial1"), []byte("Host server1\n    User admin"), 0644); err != nil {
		t.Fatalf("Failed to create partial file: %v", err)
	}

	// When - run in dry-run mode
	command := NewPartialsBuildCommand(aggregateFile, partialsDir, "#")
	command.SetDryRun(true)
	err := command.Run()
	
	// Then
	if err != nil {
		t.Fatalf("Dry run should not fail: %v", err)
	}

	// File should be unchanged in dry-run mode
	actualContent, err := os.ReadFile(aggregateFile)
	if err != nil {
		t.Fatalf("Failed to read file after dry run: %v", err)
	}
	
	if string(actualContent) != originalContent {
		t.Errorf("File was modified in dry-run mode!\nExpected: %q\nActual: %q", 
			originalContent, string(actualContent))
	}
}

