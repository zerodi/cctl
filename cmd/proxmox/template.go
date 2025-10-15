package proxmox

import "github.com/spf13/cobra"

func createTemplateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create-template",
		Short: "Create a Proxmox VM from template JSON and convert it into a template",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromConfig()
			if err != nil {
				return err
			}
			return client.CreateTemplate(cmd.Context())
		},
	}
}
