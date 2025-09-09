package src

import (
	"path/filepath"
	"strings"
)

// CommentStyle represents different comment formats
type CommentStyle struct {
	Start string
	End   string
}

// Predefined comment styles for different file types
var commentStyles = map[string]CommentStyle{
	"#":     {Start: "#", End: ""},          // Shell, Python, YAML, etc.
	"//":    {Start: "//", End: ""},         // Go, JavaScript, C++, etc.
	"--":    {Start: "--", End: ""},         // SQL, Lua, Haskell, etc.
	"/*":    {Start: "/*", End: "*/"},       // C, CSS, etc.
	";":     {Start: ";", End: ""},          // Lisp, INI files, etc.
	"%":     {Start: "%", End: ""},          // LaTeX, Erlang, etc.
	"<!--":  {Start: "<!--", End: "-->"},    // HTML, XML, etc.
	"'":     {Start: "'", End: ""},          // VB, some config files
	"rem":   {Start: "rem", End: ""},        // Batch files
	"::":    {Start: "::", End: ""},         // Batch files (alternate)
}

// File extension to comment style mapping for auto-detection
var extensionToStyle = map[string]string{
	".sh":     "#",
	".bash":   "#",
	".zsh":    "#",
	".py":     "#",
	".yml":    "#",
	".yaml":   "#",
	".conf":   "#",
	".config": "#",
	".go":     "//",
	".js":     "//",
	".ts":     "//",
	".cpp":    "//",
	".c":      "//",
	".h":      "//",
	".java":   "//",
	".cs":     "//",
	".php":    "//",
	".sql":    "--",
	".lua":    "--",
	".hs":     "--",
	".css":    "/*",
	".scss":   "//",
	".less":   "//",
	".lisp":   ";",
	".ini":    ";",
	".tex":    "%",
	".html":   "<!--",
	".xml":    "<!--",
	".vb":     "'",
	".bat":    "rem",
	".cmd":    "rem",
}

// DetectCommentStyle attempts to detect the appropriate comment style for a file
func DetectCommentStyle(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if style, exists := extensionToStyle[ext]; exists {
		return style
	}
	// Default to hash comments if no extension match
	return "#"
}

// ResolveCommentStyle resolves the comment style based on input
// Supports predefined styles, auto-detection (using "auto"), or custom characters
func ResolveCommentStyle(input string, aggregateFile string) CommentStyle {
	// Handle auto-detection
	if input == "auto" {
		detectedStyle := DetectCommentStyle(aggregateFile)
		if style, exists := commentStyles[detectedStyle]; exists {
			return style
		}
	}
	
	// Check if it's a predefined style
	if style, exists := commentStyles[input]; exists {
		return style
	}
	
	// Treat as custom comment character(s) - backward compatibility
	return CommentStyle{Start: input, End: ""}
}