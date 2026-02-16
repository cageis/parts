package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const manifestTemplate = `# Parts manifest — manages dotfiles from this directory
# Docs: https://github.com/cageis/parts

# Default settings applied to all targets (can be overridden per-target)
defaults:
  comment: "auto"    # auto-detect comment style from file extension
  backup: false      # create .bak files before modifying targets
  # mode: merge      # 'merge' (default) or 'own'

# Each target defines a file to manage
targets:
  # Example: merge SSH config partials into ~/.ssh/config
  # ssh:
  #   target: ~/.ssh/config
  #   partials: ./ssh/
  #   comment: "#"
  #   mode: merge      # preserves content outside PARTIALS markers

  # Example: fully manage ~/.vimrc from partials
  # vimrc:
  #   target: ~/.vimrc
  #   partials: ./vim/
  #   mode: own         # entire file is written from partials
`

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Generate a skeleton .parts.yaml manifest",
		Long:  `Creates a .parts.yaml file in the current directory with commented examples.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := os.Stat(".parts.yaml"); err == nil {
				return fmt.Errorf(".parts.yaml already exists in this directory")
			}

			if err := os.WriteFile(".parts.yaml", []byte(manifestTemplate), 0644); err != nil {
				return fmt.Errorf("failed to create .parts.yaml: %w", err)
			}

			fmt.Println("Created .parts.yaml — edit it to define your targets")
			return nil
		},
	}
}
