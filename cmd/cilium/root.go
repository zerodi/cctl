package cilium

import "github.com/spf13/cobra"

// New returns the `cilium` command group.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cilium",
		Short: "Operations for installing and managing Cilium",
	}
	cmd.AddCommand(installCmd())
	return cmd
}
