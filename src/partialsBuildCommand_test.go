package src

import (
	"bytes"
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
	if err := os.WriteFile(filepath.Join(partialsDir, "partials1"), []byte("Partial 1"), 0600); err != nil {
		t.Fatalf("Failed to create partial file 1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "partials2"), []byte("Partial 2"), 0600); err != nil {
		t.Fatalf("Failed to create partial file 2: %v", err)
	}
	if err := os.WriteFile(aggregateFile, []byte{}, 0600); err != nil {
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
	expected := "\n# ============================\n# PARTIALS>>>>>\n# ============================\n" +
		"Partial 1\nPartial 2\n" +
		"# ============================\n# PARTIALS<<<<<\n# ============================\n"

	if string(actual) != expected {
		t.Errorf("Expected:\n%s\nActual:\n%s", expected, string(actual))
	}
}

// testSetup creates a test environment with aggregate file and partial files
func testSetup(t *testing.T) (aggregateFile string, partialsDir string, command PartialsBuildCommand) {
	dir := t.TempDir()
	partialsDir = filepath.Join(dir, "partials")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed to create partials directory: %v", err)
	}
	aggregateFile = filepath.Join(dir, "agg")

	originalContent := "# Original SSH config\nHost example.com\n    User test\n"
	if err := os.WriteFile(aggregateFile, []byte(originalContent), 0600); err != nil {
		t.Fatalf("Failed to create aggregate file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "partial1"), []byte("Host server1\n    User admin"), 0600); err != nil {
		t.Fatalf("Failed to create partial file 1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "partial2"), []byte("Host server2\n    User root"), 0600); err != nil {
		t.Fatalf("Failed to create partial file 2: %v", err)
	}

	command = NewPartialsBuildCommand(aggregateFile, partialsDir, "#")
	return
}

// runCommandAndReadResult runs the command and returns the file content and length
func runCommandAndReadResult(t *testing.T, command PartialsBuildCommand, aggregateFile, runName string) (result []byte, length int) {
	if err := command.Run(); err != nil {
		t.Fatalf("%s run failed: %v", runName, err)
	}
	var err error
	result, err = os.ReadFile(aggregateFile)
	if err != nil {
		t.Fatalf("Failed to read %s result: %v", runName, err)
	}
	length = len(result)
	return
}

// verifyIdempotency checks that multiple runs produce identical results
func verifyIdempotency(t *testing.T, first, second, third []byte, firstLen, secondLen, thirdLen int) {
	if !bytes.Equal(first, second) {
		t.Errorf("First run differs from second run:\nFirst: %q\nSecond: %q", string(first), string(second))
	}
	if !bytes.Equal(second, third) {
		t.Errorf("Second run differs from third run:\nSecond: %q\nThird: %q", string(second), string(third))
	}
	if firstLen != secondLen || secondLen != thirdLen {
		t.Errorf("File length growing: %d -> %d -> %d", firstLen, secondLen, thirdLen)
	}
}

func TestPartialsBuildCommand_IdempotentEditing(t *testing.T) {
	// Given - setup with existing content
	aggregateFile, partialsDir, command := testSetup(t)

	// When - run command multiple times
	firstResult, firstLength := runCommandAndReadResult(t, command, aggregateFile, "First")
	secondResult, secondLength := runCommandAndReadResult(t, command, aggregateFile, "Second")
	thirdResult, thirdLength := runCommandAndReadResult(t, command, aggregateFile, "Third")

	// Then - all results should be identical (no incremental changes)
	verifyIdempotency(t, firstResult, secondResult, thirdResult, firstLength, secondLength, thirdLength)

	// Test adding new partial files
	if err := os.WriteFile(filepath.Join(partialsDir, "partial3"), []byte("Host server3\n    User guest"), 0600); err != nil {
		t.Fatalf("Failed to create partial file 3: %v", err)
	}

	// Run with new partial and verify still idempotent
	resultWithNewPartial, lengthWithNewPartial := runCommandAndReadResult(t, command, aggregateFile, "With new partial")
	finalResult, finalLength := runCommandAndReadResult(t, command, aggregateFile, "Final")

	if !bytes.Equal(resultWithNewPartial, finalResult) {
		t.Errorf("Results differ after adding new partial file:\nBefore: %q\nAfter: %q",
			string(resultWithNewPartial), string(finalResult))
	}

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
		if err := os.WriteFile(aggregateFile, []byte("test"), 0600); err != nil {
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
	if err := os.WriteFile(aggregateFile, []byte(originalContent), 0600); err != nil {
		t.Fatalf("Failed to create aggregate file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "work"), []byte("Host work\n    User admin"), 0600); err != nil {
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
	if err := os.WriteFile(aggregateFile, []byte(originalContent), 0600); err != nil {
		t.Fatalf("Failed to create aggregate file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "partial1"), []byte("Host server1\n    User admin"), 0600); err != nil {
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
