package cmd

import (
	"fmt"
	"os"

	"parts/src"

	"github.com/spf13/cobra"
)

var (
	dryRun bool
	remove bool

	rootCmd = &cobra.Command{
		Use:   "parts [flags] <aggregate-file> [partials-directory] <comment-style>",
		Short: "Merge partial configuration files into an aggregate file or remove partials sections",
		Long: `Parts is a CLI tool for merging partial configuration files into a single aggregate file.
It's designed primarily for managing SSH configurations by combining multiple partial config files.

The tool uses comment markers to manage the merged section, ensuring idempotent operations
that can be run multiple times without causing file bloat or whitespace growth.

Operations:
  - Build mode (default): Merge partials into the aggregate file
  - Remove mode (--remove): Remove the partials section from the aggregate file

Comment Style Support:
  The comment-style parameter ensures the section markers blend seamlessly with your file type:
  - Use language-appropriate comments: "#" for SSH/shell, "//" for JavaScript, "/*" for CSS, etc.
  - Auto-detection: "auto" (automatically picks the right style based on file extension)
  - Predefined styles: "#", "//", "--", "/*", ";", "%", "<!--", "'", "rem", "::"
  - Custom characters: Any string (for special cases or backward compatibility)
  
  Why different comment styles? The markers need to be valid comments in your target file
  so they don't interfere with syntax highlighting, parsing, or execution.`,
		Example: `  # Build mode: Merge partials into aggregate file
  parts ~/.ssh/config ~/.ssh/config.d "#"
  parts app.js ./partials "//"
  parts schema.sql ./sql-partials "auto"
  parts --dry-run ~/.ssh/config ~/.ssh/config.d "#"
  
  # Remove mode: Remove partials section from file
  parts --remove ~/.ssh/config "#"
  parts --remove styles.css "/*"
  parts --remove config.py "auto"
  
  # Auto-detection works great for most file types
  parts config.py ./python-configs "auto"`,
		Args: func(cmd *cobra.Command, args []string) error {
			if remove {
				// Remove mode: requires 2 args (file and comment-style)
				if len(args) != 2 {
					return fmt.Errorf("remove mode requires exactly 2 arguments: <aggregate-file> <comment-style>, got %d", len(args))
				}
			} else {
				// Build mode: requires 3 args (file, partials-dir, comment-style)
				if len(args) != 3 {
					return fmt.Errorf("build mode requires exactly 3 arguments: <aggregate-file> <partials-directory> <comment-style>, got %d", len(args))
				}
			}
			return nil
		},
		RunE: runParts,
	}
)

func runParts(cmd *cobra.Command, args []string) error {
	if remove {
		// Remove mode: parts --remove <aggregate-file> <comment-style>
		aggregateFile := args[0]
		commentStyle := args[1]

		command := src.NewPartialsRemoveCommand(aggregateFile, commentStyle)
		command.SetDryRun(dryRun)

		return command.Run()
	} else {
		// Build mode: parts <aggregate-file> <partials-directory> <comment-style>
		aggregateFile := args[0]
		partialsDir := args[1]
		commentStyle := args[2]

		command := src.NewPartialsBuildCommand(aggregateFile, partialsDir, commentStyle)
		command.SetDryRun(dryRun)

		return command.Run()
	}
}

// Execute runs the root command
func Execute() {
	rootCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "preview changes without modifying files")
	rootCmd.Flags().BoolVarP(&remove, "remove", "r", false, "remove partials section from aggregate file")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
