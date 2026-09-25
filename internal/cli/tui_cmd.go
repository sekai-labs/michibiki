package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sekai-labs/michibiki/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:     "tui",
	Aliases: []string{"ui", "dashboard"},
	Short:   "Launch the interactive full-terminal user interface",
	RunE: func(cmd *cobra.Command, args []string) error {
		prov, profile, err := resolveProvider(cmd)
		if err != nil {
			return fmt.Errorf("failed to initialize device for TUI: %w", err)
		}
		defer prov.Disconnect(cmd.Context())

		return tui.Run(prov, profile.Name)
	},
}
