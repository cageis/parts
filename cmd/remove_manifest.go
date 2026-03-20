package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cageis/parts/src"
	"github.com/spf13/cobra"
)

var manifestRemovePath string

func newManifestRemoveCmd() *cobra.Command {
	var removeDryRun bool

	cmd := &cobra.Command{
		Use:   "remove [target-name...]",
		Short: "Remove managed sections from manifest targets",
		Long: `Reads .parts.yaml and removes the managed content from each target.

For 'merge' mode targets, the PARTIALS markers and their content are removed,
preserving any user content outside the markers.

For 'own' mode targets, the target file is deleted entirely.`,
		Example: `  parts remove           # Remove all targets
  parts remove ssh       # Remove only the 'ssh' target
  parts remove --dry-run # Preview what would be removed`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestPath := manifestRemovePath
			if manifestPath == "" {
				manifestPath = resolveManifestPath()
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
			for _, name := range names {
				target := manifest.ResolvedTarget(name)

				switch target.Mode {
				case "merge":
					// NewPartialsRemoveCommand handles tilde expansion internally
					rmCmd, rmErr := src.NewPartialsRemoveCommand(target.Target, target.Comment)
					if rmErr != nil {
						errors = append(errors, fmt.Errorf("target '%s': %w", name, rmErr))
						continue
					}
					rmCmd.SetDryRun(removeDryRun)
					if runErr := rmCmd.Run(); runErr != nil {
						errors = append(errors, fmt.Errorf("target '%s': %w", name, runErr))
					}

				case "own":
					expandedTarget, expandErr := src.ExpandTildePrefix(target.Target)
					if expandErr != nil {
						errors = append(errors, fmt.Errorf("target '%s': %w", name, expandErr))
						continue
					}

					if removeDryRun {
						fmt.Printf("DRY RUN: Would delete '%s' (own mode)\n", expandedTarget)
					} else {
						if err := os.Remove(expandedTarget); err != nil {
							if !os.IsNotExist(err) {
								errors = append(errors, fmt.Errorf("target '%s': failed to delete '%s': %w", name, expandedTarget, err))
							}
						} else {
							fmt.Printf("Deleted '%s' (own mode)\n", expandedTarget)
						}
					}
				}
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

	cmd.Flags().BoolVarP(&removeDryRun, "dry-run", "n", false, "preview changes without modifying files")
	return cmd
}
