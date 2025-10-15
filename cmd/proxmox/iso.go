package proxmox

import (
	"errors"

	"github.com/spf13/cobra"
)

func getTalosImageCmd() *cobra.Command {
	var version string

	cmd := &cobra.Command{
		Use:   "get-talos-image",
		Short: "Download Talos ISO for the given version and upload to Proxmox storage",
		RunE: func(cmd *cobra.Command, args []string) error {
			if version == "" {
				return errors.New("version is required (e.g. --version 1.11.2)")
			}

			client, err := clientFromConfig()
			if err != nil {
				return err
			}

			return client.GetTalosImage(cmd.Context(), version)
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Talos release version (e.g. 1.11.2)")
	return cmd
}
