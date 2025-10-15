package kind

import "github.com/spf13/cobra"

func New() *cobra.Command {
	cmd := &cobra.Command{Use: "kind", Short: "Commands for managing the kind cluster"}
	cmd.AddCommand(upCmd())
	cmd.AddCommand(downCmd())
	cmd.AddCommand(resetCmd())
	return cmd
}
