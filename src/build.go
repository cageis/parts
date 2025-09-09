package src

import (
	"fmt"
	"io/ioutil"
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

// NewPartialsBuildCommand creates a new build command
func NewPartialsBuildCommand(aggregateFile, partialsDir, commentChars string) PartialsBuildCommand {
	aggregateFile = ExpandTildePrefix(aggregateFile)
	partialsDir = ExpandTildePrefix(partialsDir)

	return PartialsBuildCommand{aggregateFile, partialsDir, commentChars, false}
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
		return fmt.Sprintf("%s\n%s PARTIALS>>>>>\n%s", style.Start, style.Start, style.End)
	}
	// Single-character comment style with header block
	return fmt.Sprintf("%s ============================\n%s PARTIALS>>>>>\n%s ============================",
		style.Start, style.Start, style.Start)
}

// GetEndFlag returns the end marker for this build command
func (p PartialsBuildCommand) GetEndFlag() string {
	style := p.getCommentStyle()
	if style.End != "" {
		// Multi-character comment style with proper footer block
		return fmt.Sprintf("%s\n%s PARTIALS<<<<<\n%s", style.Start, style.Start, style.End)
	}
	// Single-character comment style with footer block
	return fmt.Sprintf("%s ============================\n%s PARTIALS<<<<<\n%s ============================",
		style.Start, style.Start, style.Start)
}

// Run executes the build command
func (p PartialsBuildCommand) Run() error {
	path, err := filepath.Abs(p.aggregateFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for aggregate file '%s': %w", p.aggregateFile, err)
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

	files, err := ioutil.ReadDir(p.partialsDir)
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

	err = ioutil.WriteFile(p.aggregateFile, []byte(output), 0600)
	if err != nil {
		return fmt.Errorf("failed to write aggregate file '%s': %w", p.aggregateFile, err)
	}

	return nil
}
