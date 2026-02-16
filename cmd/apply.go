package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/cageis/parts/src"
	"github.com/spf13/cobra"
)

// applyManifestPath allows tests to override the manifest location
var applyManifestPath string

func newApplyCmd() *cobra.Command {
	var applyDryRun bool

	cmd := &cobra.Command{
		Use:   "apply [target-name...]",
		Short: "Apply manifest targets — merge partials into target files",
		Long: `Reads .parts.yaml from the current directory and applies each target.

For 'merge' mode targets, partials are merged into the target file between
PARTIALS markers (existing file content outside the markers is preserved).

For 'own' mode targets, the target file is entirely written from the
concatenated partials (the file is fully managed by Parts).

If target names are specified, only those targets are applied.
If no target names are specified, all targets are applied.`,
		Example: `  parts apply            # Apply all targets
  parts apply ssh        # Apply only the 'ssh' target
  parts apply --dry-run  # Preview changes without modifying files`,
		RunE: func(cmd *cobra.Command, args []string) error {
			manifestPath := applyManifestPath
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
			for _, name := range names {
				target := manifest.ResolvedTarget(name)

				switch target.Mode {
				case "merge":
					// NewPartialsBuildCommand handles tilde expansion internally
					buildCmd, buildErr := src.NewPartialsBuildCommand(target.Target, target.Partials, target.Comment)
					if buildErr != nil {
						errors = append(errors, fmt.Errorf("target '%s': %w", name, buildErr))
						continue
					}
					buildCmd.SetDryRun(applyDryRun)
					if runErr := buildCmd.Run(); runErr != nil {
						errors = append(errors, fmt.Errorf("target '%s': %w", name, runErr))
					}

				case "own":
					// Own mode needs manual tilde expansion
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

					ownCmd := src.NewPartialsOwnCommand(expandedTarget, expandedPartials, target.Comment)
					ownCmd.SetDryRun(applyDryRun)
					if runErr := ownCmd.Run(); runErr != nil {
						errors = append(errors, fmt.Errorf("target '%s': %w", name, runErr))
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

	cmd.Flags().BoolVarP(&applyDryRun, "dry-run", "n", false, "preview changes without modifying files")
	return cmd
}
