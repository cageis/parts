package src

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveCommentStyle_PredefinedStyles(t *testing.T) {
	tests := []struct {
		input    string
		expected CommentStyle
	}{
		{"#", CommentStyle{Start: "#", End: ""}},
		{"//", CommentStyle{Start: "//", End: ""}},
		{"--", CommentStyle{Start: "--", End: ""}},
		{"/*", CommentStyle{Start: "/*", End: "*/"}},
		{";", CommentStyle{Start: ";", End: ""}},
		{"%", CommentStyle{Start: "%", End: ""}},
		{"<!--", CommentStyle{Start: "<!--", End: "-->"}},
		{"'", CommentStyle{Start: "'", End: ""}},
		{"rem", CommentStyle{Start: "rem", End: ""}},
		{"::", CommentStyle{Start: "::", End: ""}},
	}

	for _, test := range tests {
		result := ResolveCommentStyle(test.input, "test.txt")
		if result.Start != test.expected.Start || result.End != test.expected.End {
			t.Errorf("For input %q, expected %+v, got %+v", test.input, test.expected, result)
		}
	}
}

func TestResolveCommentStyle_CustomCharacters(t *testing.T) {
	tests := []struct {
		input    string
		expected CommentStyle
	}{
		{"@", CommentStyle{Start: "@", End: ""}},
		{"###", CommentStyle{Start: "###", End: ""}},
		{">>", CommentStyle{Start: ">>", End: ""}},
	}

	for _, test := range tests {
		result := ResolveCommentStyle(test.input, "test.txt")
		if result.Start != test.expected.Start || result.End != test.expected.End {
			t.Errorf("For input %q, expected %+v, got %+v", test.input, test.expected, result)
		}
	}
}

func TestDetectCommentStyle_FileExtensions(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"config.sh", "#"},
		{"script.py", "#"},
		{"config.yml", "#"},
		{"app.js", "//"},
		{"main.go", "//"},
		{"style.css", "/*"},
		{"query.sql", "--"},
		{"init.lua", "--"},
		{"config.ini", ";"},
		{"document.tex", "%"},
		{"page.html", "<!--"},
		{"data.xml", "<!--"},
		{"script.bat", "rem"},
		{"unknown.xyz", "#"}, // default
	}

	for _, test := range tests {
		result := DetectCommentStyle(test.filename)
		if result != test.expected {
			t.Errorf("For filename %q, expected %q, got %q", test.filename, test.expected, result)
		}
	}
}

func TestResolveCommentStyle_AutoDetection(t *testing.T) {
	tests := []struct {
		filename string
		expected CommentStyle
	}{
		{"config.py", CommentStyle{Start: "#", End: ""}},
		{"app.js", CommentStyle{Start: "//", End: ""}},
		{"styles.css", CommentStyle{Start: "/*", End: "*/"}},
		{"query.sql", CommentStyle{Start: "--", End: ""}},
	}

	for _, test := range tests {
		result := ResolveCommentStyle("auto", test.filename)
		if result.Start != test.expected.Start || result.End != test.expected.End {
			t.Errorf("For auto detection with %q, expected %+v, got %+v", test.filename, test.expected, result)
		}
	}
}

func TestPartialsBuildCommand_DifferentCommentStyles(t *testing.T) {
	tests := []struct {
		name          string
		commentStyle  string
		expectedStart string
		expectedEnd   string
	}{
		{"Hash comments", "#", "# ============================\n# PARTIALS>>>>>\n# ============================", "# ============================\n# PARTIALS<<<<<\n# ============================"},
		{"Slash comments", "//", "// ============================\n// PARTIALS>>>>>\n// ============================", "// ============================\n// PARTIALS<<<<<\n// ============================"},
		{"Dash comments", "--", "-- ============================\n-- PARTIALS>>>>>\n-- ============================", "-- ============================\n-- PARTIALS<<<<<\n-- ============================"},
		{"Block comments", "/*", "/*\n/* PARTIALS>>>>>\n*/", "/*\n/* PARTIALS<<<<<\n*/"},
		{"HTML comments", "<!--", "<!--\n<!-- PARTIALS>>>>>\n-->", "<!--\n<!-- PARTIALS<<<<<\n-->"},
		{"Custom characters", "@@", "@@ ============================\n@@ PARTIALS>>>>>\n@@ ============================", "@@ ============================\n@@ PARTIALS<<<<<\n@@ ============================"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			partialsDir := filepath.Join(dir, "partials")
			if err := os.MkdirAll(partialsDir, 0755); err != nil {
				t.Fatalf("Failed to create partials directory: %v", err)
			}
			aggregateFile := filepath.Join(dir, "agg")

			// Create test files
			if err := os.WriteFile(aggregateFile, []byte("Original content\n"), 0600); err != nil {
				t.Fatalf("Failed to create aggregate file: %v", err)
			}
			if err := os.WriteFile(filepath.Join(partialsDir, "partial1"), []byte("Test content"), 0600); err != nil {
				t.Fatalf("Failed to create partial file: %v", err)
			}

			// Test the command
			command := NewPartialsBuildCommand(aggregateFile, partialsDir, test.commentStyle)

			// Check marker generation
			if command.GetStartFlag() != test.expectedStart {
				t.Errorf("Expected start flag %q, got %q", test.expectedStart, command.GetStartFlag())
			}
			if command.GetEndFlag() != test.expectedEnd {
				t.Errorf("Expected end flag %q, got %q", test.expectedEnd, command.GetEndFlag())
			}

			// Run the command
			if err := command.Run(); err != nil {
				t.Fatalf("Command failed: %v", err)
			}

			// Verify the output contains the expected markers
			result, err := os.ReadFile(aggregateFile)
			if err != nil {
				t.Fatalf("Failed to read result: %v", err)
			}
			resultStr := string(result)

			if !strings.Contains(resultStr, test.expectedStart) {
				t.Errorf("Result should contain start marker %q", test.expectedStart)
			}
			if !strings.Contains(resultStr, test.expectedEnd) {
				t.Errorf("Result should contain end marker %q", test.expectedEnd)
			}
		})
	}
}

func TestPartialsBuildCommand_AutoDetection(t *testing.T) {
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	if err := os.MkdirAll(partialsDir, 0755); err != nil {
		t.Fatalf("Failed to create partials directory: %v", err)
	}

	// Test with a Python file
	aggregateFile := filepath.Join(dir, "config.py")
	if err := os.WriteFile(aggregateFile, []byte("# Original Python config\n"), 0644); err != nil {
		t.Fatalf("Failed to create aggregate file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "partial1"), []byte("DEBUG = True"), 0644); err != nil {
		t.Fatalf("Failed to create partial file: %v", err)
	}

	// Use auto-detection
	command := NewPartialsBuildCommand(aggregateFile, partialsDir, "auto")
	if err := command.Run(); err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Verify it used hash comments
	result, err := os.ReadFile(aggregateFile)
	if err != nil {
		t.Fatalf("Failed to read result: %v", err)
	}
	resultStr := string(result)

	if !strings.Contains(resultStr, "# PARTIALS>>>>>") {
		t.Error("Auto-detection should have used hash comments for .py file")
	}
	if !strings.Contains(resultStr, "# PARTIALS<<<<<") {
		t.Error("Auto-detection should have used hash comments for .py file")
	}
	if !strings.Contains(resultStr, "# ============================") {
		t.Error("Auto-detection should have used enhanced comment blocks for .py file")
	}
}
