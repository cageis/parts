package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/cageis/parts/src"
	"github.com/spf13/cobra"
)

var syncManifestPath string

func newSyncCmd() *cobra.Command {
	var syncDryRun bool

	cmd := &cobra.Command{
		Use:   "sync [target-name...]",
		Short: "Sync changes from target files back into partials",
		Long: `Reads .parts.yaml and detects changes in target files, pulling modified
content back into the partial source files.

Uses the '# Source: <path>' comments to map content back to individual
partial files.`,
		Example: `  parts sync            # Sync all targets
  parts sync ssh        # Sync only the 'ssh' target
  parts sync --dry-run  # Preview what would be synced`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestPath := syncManifestPath
			if manifestPath == "" {
				manifestPath = ".parts.yaml"
			}

			absManifest, err := filepath.Abs(manifestPath)
			if err != nil {
				return fmt.Errorf("failed to resolve manifest path: %w", err)
			}

			manifest, err := src.LoadManifest(absManifest)
			if err != nil {
				return err
			}

			names, err := manifest.FilterTargets(args)
			if err != nil {
				return err
			}

			var errors []error
			totalUpdated := 0

			for _, name := range names {
				target := manifest.ResolvedTarget(name)

				expandedTarget, expandErr := src.ExpandTildePrefix(target.Target)
				if expandErr != nil {
					errors = append(errors, fmt.Errorf("target '%s': %w", name, expandErr))
					continue
				}

				expandedPartials, expandErr := src.ExpandTildePrefix(target.Partials)
				if expandErr != nil {
					errors = append(errors, fmt.Errorf("target '%s': %w", name, expandErr))
					continue
				}

				result, syncErr := src.SyncTarget(expandedTarget, expandedPartials, target.Comment, target.Mode, syncDryRun)
				if syncErr != nil {
					errors = append(errors, fmt.Errorf("target '%s': %w", name, syncErr))
					continue
				}

				totalUpdated += result.UpdatedFiles
			}

			if syncDryRun {
				fmt.Printf("DRY RUN: %d partial file(s) would be updated\n", totalUpdated)
			} else if totalUpdated == 0 {
				fmt.Println("All partials are in sync")
			}

			if len(errors) > 0 {
				for _, e := range errors {
					fmt.Fprintf(cmd.ErrOrStderr(), "Error: %v\n", e)
				}
				return fmt.Errorf("%d target(s) failed", len(errors))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&syncDryRun, "dry-run", "n", false, "preview changes without modifying files")
	return cmd
}
