package src

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// PartialsOwnCommand handles writing entire files from partials (no markers)
type PartialsOwnCommand struct {
	targetFile   string
	partialsDir  string
	commentChars string
	dryRun       bool
}

// NewPartialsOwnCommand creates a new own command.
// commentChars is optional — if non-empty, source-path comments are added.
func NewPartialsOwnCommand(targetFile, partialsDir, commentChars string) PartialsOwnCommand {
	return PartialsOwnCommand{
		targetFile:   targetFile,
		partialsDir:  partialsDir,
		commentChars: commentChars,
	}
}

// SetDryRun sets the dry-run mode
func (p *PartialsOwnCommand) SetDryRun(dryRun bool) {
	p.dryRun = dryRun
}

// Run executes the own command
func (p PartialsOwnCommand) Run() error {
	// Get original file permissions if file exists
	var originalMode fs.FileMode = 0644
	if info, err := os.Stat(p.targetFile); err == nil {
		originalMode = info.Mode()
	}

	files, err := os.ReadDir(p.partialsDir)
	if err != nil {
		return fmt.Errorf("failed to read partials directory '%s': %w", p.partialsDir, err)
	}

	var output strings.Builder

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		partialPath := filepath.Join(p.partialsDir, file.Name())
		content, readErr := os.ReadFile(partialPath)
		if readErr != nil {
			return fmt.Errorf("failed to read partial file '%s': %w", partialPath, readErr)
		}

		// Add source comment if comment style is provided
		if p.commentChars != "" {
			style := ResolveCommentStyle(p.commentChars, p.targetFile)
			if style.End != "" {
				output.WriteString(fmt.Sprintf("%s Source: %s %s\n", style.Start, partialPath, style.End))
			} else {
				output.WriteString(fmt.Sprintf("%s Source: %s\n", style.Start, partialPath))
			}
		}

		output.Write(content)
		// Ensure each partial ends with a newline
		if len(content) > 0 && content[len(content)-1] != '\n' {
			output.WriteByte('\n')
		}
	}

	if p.dryRun {
		fmt.Printf("DRY RUN: Would write to '%s' (own mode)\n", p.targetFile)
		fmt.Printf("Content preview:\n")
		fmt.Printf("--- BEGIN FILE CONTENT ---\n")
		fmt.Print(output.String())
		fmt.Printf("--- END FILE CONTENT ---\n")
		fmt.Printf("Total length: %d characters\n", output.Len())
		return nil
	}

	// Create target directory if it doesn't exist
	targetDir := filepath.Dir(p.targetFile)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory '%s': %w", targetDir, err)
	}

	if err := os.WriteFile(p.targetFile, []byte(output.String()), originalMode); err != nil {
		return fmt.Errorf("failed to write target file '%s': %w", p.targetFile, err)
	}

	// Count partial files for success message
	partialCount := 0
	for _, f := range files {
		if !f.IsDir() {
			partialCount++
		}
	}
	fmt.Printf("Wrote %d partial(s) to '%s' (own mode)\n", partialCount, p.targetFile)

	return nil
}
