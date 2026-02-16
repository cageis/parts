package src

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// SyncResult reports what the sync operation found/did
type SyncResult struct {
	UpdatedFiles int
	SkippedFiles int
	ChangedPaths []string
}

// ExtractPartialSections parses file content and splits it by "# Source: <path>" comments.
// Returns a map of source-path -> content-after-that-comment.
func ExtractPartialSections(content, commentChars string) (map[string]string, error) {
	style := ResolveCommentStyle(commentChars, "")
	prefix := fmt.Sprintf("%s Source: ", style.Start)

	lines := strings.Split(content, "\n")
	sections := make(map[string]string)
	var currentPath string
	var currentContent strings.Builder

	for _, line := range lines {
		// Check for source comment
		if strings.HasPrefix(line, prefix) {
			// Save previous section
			if currentPath != "" {
				sections[currentPath] = normalizeSectionContent(currentContent.String())
			}
			// Extract path from source comment
			pathPart := strings.TrimPrefix(line, prefix)
			// Remove closing comment chars if present (e.g., " */")
			if style.End != "" {
				pathPart = strings.TrimSuffix(pathPart, " "+style.End)
			}
			currentPath = strings.TrimSpace(pathPart)
			currentContent.Reset()
			continue
		}

		// Skip marker lines (PARTIALS>>>>>, PARTIALS<<<<<, separator)
		if strings.Contains(line, PartialStartMarker) ||
			strings.Contains(line, PartialEndMarker) ||
			strings.Contains(line, MarkerSeparator) {
			continue
		}

		if currentPath != "" {
			currentContent.WriteString(line)
			currentContent.WriteByte('\n')
		}
	}

	// Save last section
	if currentPath != "" {
		sections[currentPath] = normalizeSectionContent(currentContent.String())
	}

	return sections, nil
}

// SyncTarget reads the target file, extracts sections by source comment,
// and writes changed content back to the partial files.
func SyncTarget(targetFile, partialsDir, commentChars, mode string, dryRun bool) (*SyncResult, error) {
	content, err := os.ReadFile(targetFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read target file '%s': %w", targetFile, err)
	}

	var sectionContent string
	contentStr := string(content)

	if mode == "merge" {
		// Extract only the content between PARTIALS markers
		style := ResolveCommentStyle(commentChars, targetFile)
		startFlag := buildStartFlag(style)
		endFlag := buildEndFlag(style)

		startIdx := strings.Index(contentStr, startFlag)
		endIdx := strings.Index(contentStr, endFlag)
		if startIdx == -1 || endIdx == -1 {
			return &SyncResult{}, nil // No managed section found
		}
		sectionContent = contentStr[startIdx:endIdx]
	} else {
		// Own mode: entire file is managed
		sectionContent = contentStr
	}

	sections, err := ExtractPartialSections(sectionContent, commentChars)
	if err != nil {
		return nil, fmt.Errorf("failed to extract sections: %w", err)
	}

	result := &SyncResult{}

	for sourcePath, newContent := range sections {
		// Verify the source path is within the partials directory
		absSource, _ := filepath.Abs(sourcePath)
		absPartials, _ := filepath.Abs(partialsDir)
		if !strings.HasPrefix(absSource, absPartials) {
			result.SkippedFiles++
			continue
		}

		// Read current partial content
		existing, readErr := os.ReadFile(sourcePath)
		if readErr != nil {
			result.SkippedFiles++
			continue
		}

		// Compare — trim trailing newlines for comparison to handle minor formatting diffs
		existingTrimmed := strings.TrimRight(string(existing), "\n")
		newTrimmed := strings.TrimRight(newContent, "\n")

		if existingTrimmed == newTrimmed {
			continue // No change
		}

		result.UpdatedFiles++
		result.ChangedPaths = append(result.ChangedPaths, sourcePath)

		if dryRun {
			fmt.Printf("DRY RUN: Would update '%s'\n", sourcePath)
			continue
		}

		// Write back — preserve trailing newline
		writeContent := strings.TrimRight(newContent, "\n") + "\n"
		if writeErr := os.WriteFile(sourcePath, []byte(writeContent), 0644); writeErr != nil {
			return nil, fmt.Errorf("failed to write partial '%s': %w", sourcePath, writeErr)
		}
		fmt.Printf("Updated '%s'\n", sourcePath)
	}

	return result, nil
}

// normalizeSectionContent trims extra trailing newlines from split artifacts
// and ensures content ends with exactly one newline.
func normalizeSectionContent(s string) string {
	s = strings.TrimRight(s, "\n")
	if s != "" {
		s += "\n"
	}
	return s
}

// buildStartFlag constructs the start marker string for a given comment style
func buildStartFlag(style CommentStyle) string {
	if style.End != "" {
		return fmt.Sprintf("%s\n%s %s\n%s", style.Start, style.Start, PartialStartMarker, style.End)
	}
	return fmt.Sprintf("%s %s\n%s %s\n%s %s",
		style.Start, MarkerSeparator, style.Start, PartialStartMarker, style.Start, MarkerSeparator)
}

// buildEndFlag constructs the end marker string for a given comment style
func buildEndFlag(style CommentStyle) string {
	if style.End != "" {
		return fmt.Sprintf("%s\n%s %s\n%s", style.Start, style.Start, PartialEndMarker, style.End)
	}
	return fmt.Sprintf("%s %s\n%s %s\n%s %s",
		style.Start, MarkerSeparator, style.Start, PartialEndMarker, style.Start, MarkerSeparator)
}
