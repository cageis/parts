package src

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// PartialsBuildCommand handles building/merging partials into aggregate files
type PartialsBuildCommand struct {
	aggregateFile string
	partialsDir   string
	commentChars  string
	dryRun        bool
}

// NewPartialsBuildCommand creates a new build command.
// Returns an error if path expansion fails.
func NewPartialsBuildCommand(aggregateFile, partialsDir, commentChars string) (PartialsBuildCommand, error) {
	expandedAgg, err := ExpandTildePrefix(aggregateFile)
	if err != nil {
		return PartialsBuildCommand{}, fmt.Errorf("failed to expand aggregate file path: %w", err)
	}
	expandedPartials, err := ExpandTildePrefix(partialsDir)
	if err != nil {
		return PartialsBuildCommand{}, fmt.Errorf("failed to expand partials directory path: %w", err)
	}

	return PartialsBuildCommand{expandedAgg, expandedPartials, commentChars, false}, nil
}

// SetDryRun sets the dry-run mode for the build command
func (p *PartialsBuildCommand) SetDryRun(dryRun bool) {
	p.dryRun = dryRun
}

// getCommentStyle returns the resolved comment style for this command
func (p PartialsBuildCommand) getCommentStyle() CommentStyle {
	return ResolveCommentStyle(p.commentChars, p.aggregateFile)
}

// GetStartFlag returns the start marker for this build command
func (p PartialsBuildCommand) GetStartFlag() string {
	style := p.getCommentStyle()
	if style.End != "" {
		// Multi-character comment style with proper header block
		return fmt.Sprintf("%s\n%s %s\n%s", style.Start, style.Start, PartialStartMarker, style.End)
	}
	// Single-character comment style with header block
	return fmt.Sprintf("%s %s\n%s %s\n%s %s",
		style.Start, MarkerSeparator, style.Start, PartialStartMarker, style.Start, MarkerSeparator)
}

// GetEndFlag returns the end marker for this build command
func (p PartialsBuildCommand) GetEndFlag() string {
	style := p.getCommentStyle()
	if style.End != "" {
		// Multi-character comment style with proper footer block
		return fmt.Sprintf("%s\n%s %s\n%s", style.Start, style.Start, PartialEndMarker, style.End)
	}
	// Single-character comment style with footer block
	return fmt.Sprintf("%s %s\n%s %s\n%s %s",
		style.Start, MarkerSeparator, style.Start, PartialEndMarker, style.Start, MarkerSeparator)
}

// Run executes the build command
func (p PartialsBuildCommand) Run() error {
	path, err := filepath.Abs(p.aggregateFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for aggregate file '%s': %w", p.aggregateFile, err)
	}

	// Get original file permissions before reading
	var originalMode fs.FileMode = 0600 // default if file doesn't exist
	if info, statErr := os.Stat(path); statErr == nil {
		originalMode = info.Mode()
	}

	agg, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read aggregate file '%s': %w", path, err)
	}
	output := string(agg)

	startIndex := strings.Index(output, p.GetStartFlag())
	endIndex := strings.Index(output, p.GetEndFlag())

	if startIndex != -1 && endIndex != -1 {
		before := output[:startIndex]
		afterStart := endIndex + len(p.GetEndFlag())
		// Skip the trailing newline after the end flag if present
		if afterStart < len(output) && output[afterStart] == '\n' {
			afterStart++
		}
		after := output[afterStart:]
		output = before + after
	}

	// Add separator newline if content doesn't end with one
	if !strings.HasSuffix(output, "\n") {
		output += "\n"
	}
	output += p.GetStartFlag()
	output += "\n"

	files, err := os.ReadDir(p.partialsDir)
	if err != nil {
		return fmt.Errorf("failed to read partials directory '%s': %w", p.partialsDir, err)
	}

	// Each file: read contents into var to be written later.
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		partialPath := filepath.Join(p.partialsDir, file.Name())
		fileContents, readErr := os.ReadFile(partialPath)
		if readErr != nil {
			return fmt.Errorf("failed to read partial file '%s': %w", partialPath, readErr)
		}

		// Add source file path comment before each partial's content
		style := p.getCommentStyle()
		if style.End != "" {
			// Multi-character comment style - need to close the comment
			output += fmt.Sprintf("%s Source: %s %s\n", style.Start, partialPath, style.End)
		} else {
			// Single-character comment style
			output += fmt.Sprintf("%s Source: %s\n", style.Start, partialPath)
		}
		output += string(fileContents)
		output += "\n"
	}

	output += p.GetEndFlag()
	output += "\n"

	if p.dryRun {
		fmt.Printf("DRY RUN: Would write to '%s'\n", p.aggregateFile)
		fmt.Printf("Content preview:\n")
		fmt.Printf("--- BEGIN FILE CONTENT ---\n")
		fmt.Print(output)
		fmt.Printf("--- END FILE CONTENT ---\n")
		fmt.Printf("Total length: %d characters\n", len(output))
		return nil
	}

	err = os.WriteFile(p.aggregateFile, []byte(output), originalMode)
	if err != nil {
		return fmt.Errorf("failed to write aggregate file '%s': %w", p.aggregateFile, err)
	}

	// Count partial files for success message
	partialCount := 0
	for _, f := range files {
		if !f.IsDir() {
			partialCount++
		}
	}
	fmt.Printf("Merged %d partial(s) into '%s'\n", partialCount, p.aggregateFile)

	return nil
}
