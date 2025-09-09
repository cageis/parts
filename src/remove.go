package src

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PartialsRemoveCommand handles removing partials sections from files
type PartialsRemoveCommand struct {
	aggregateFile string
	commentChars  string
	dryRun        bool
}

// NewPartialsRemoveCommand creates a new remove command
func NewPartialsRemoveCommand(aggregateFile string, commentChars string) PartialsRemoveCommand {
	aggregateFile = ExpandTildePrefix(aggregateFile)
	return PartialsRemoveCommand{aggregateFile, commentChars, false}
}

// SetDryRun sets the dry-run mode for the remove command
func (p *PartialsRemoveCommand) SetDryRun(dryRun bool) {
	p.dryRun = dryRun
}

// getCommentStyle returns the resolved comment style for this remove command
func (p PartialsRemoveCommand) getCommentStyle() CommentStyle {
	return ResolveCommentStyle(p.commentChars, p.aggregateFile)
}

// GetStartFlag returns the start marker for this remove command
func (p PartialsRemoveCommand) GetStartFlag() string {
	style := p.getCommentStyle()
	if style.End != "" {
		// Multi-character comment style with proper header block
		return fmt.Sprintf("%s\n%s PARTIALS>>>>>\n%s", style.Start, style.Start, style.End)
	}
	// Single-character comment style with header block
	return fmt.Sprintf("%s ============================\n%s PARTIALS>>>>>\n%s ============================", style.Start, style.Start, style.Start)
}

// GetEndFlag returns the end marker for this remove command
func (p PartialsRemoveCommand) GetEndFlag() string {
	style := p.getCommentStyle()
	if style.End != "" {
		// Multi-character comment style with proper footer block
		return fmt.Sprintf("%s\n%s PARTIALS<<<<<\n%s", style.Start, style.Start, style.End)
	}
	// Single-character comment style with footer block
	return fmt.Sprintf("%s ============================\n%s PARTIALS<<<<<\n%s ============================", style.Start, style.Start, style.Start)
}

// Run executes the remove command
func (p PartialsRemoveCommand) Run() error {
	path, err := filepath.Abs(p.aggregateFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for aggregate file '%s': %w", p.aggregateFile, err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read aggregate file '%s': %w", path, err)
	}

	output := string(content)
	startIndex := strings.Index(output, p.GetStartFlag())
	endIndex := strings.Index(output, p.GetEndFlag())

	if startIndex == -1 || endIndex == -1 {
		if p.dryRun {
			fmt.Printf("DRY RUN: No partials section found in '%s' to remove\n", p.aggregateFile)
			return nil
		}
		return fmt.Errorf("no partials section found in file '%s' (looking for comment style '%s')", p.aggregateFile, p.commentChars)
	}

	// Remove the entire partials section
	before := output[:startIndex]
	afterStart := endIndex + len(p.GetEndFlag())
	// Skip the trailing newline after the end flag if present
	if afterStart < len(output) && output[afterStart] == '\n' {
		afterStart++
	}
	after := output[afterStart:]

	// Clean up any extra newlines at the end of before section
	before = strings.TrimRight(before, "\n") + "\n"

	result := before + after

	if p.dryRun {
		fmt.Printf("DRY RUN: Would remove partials section from '%s'\n", p.aggregateFile)
		fmt.Printf("Original length: %d characters\n", len(output))
		fmt.Printf("New length: %d characters\n", len(result))
		fmt.Printf("Removed %d characters\n", len(output)-len(result))
		fmt.Printf("Content preview:\n")
		fmt.Printf("--- BEGIN FILE CONTENT ---\n")
		fmt.Print(result)
		fmt.Printf("--- END FILE CONTENT ---\n")
		return nil
	}

	err = os.WriteFile(p.aggregateFile, []byte(result), 0644)
	if err != nil {
		return fmt.Errorf("failed to write aggregate file '%s': %w", p.aggregateFile, err)
	}

	fmt.Printf("✅ Removed partials section from '%s'\n", p.aggregateFile)
	return nil
}
