package secrets

import "github.com/spf13/cobra"

// New returns the secrets command group.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "Helpers for fetching Cluster API kubeconfig/talosconfig secrets",
	}
	cmd.AddCommand(getKubeconfigCmd())
	cmd.AddCommand(getTalosconfigCmd())
	cmd.AddCommand(kubeconfigViaTalosCmd())
	return cmd
}
