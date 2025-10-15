package proxmox

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func refreshSchematicCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "refresh-schematic",
		Short: "Upload Talos schematic YAML and cache returned ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromConfig()
			if err != nil {
				return err
			}

			id, err := client.RefreshSchematic(cmd.Context())
			if err != nil {
				return err
			}

			log.Info().Str("schematicID", id).Msg("Talos schematic cached")
			fmt.Fprintln(cmd.OutOrStdout(), id)
			return nil
		},
	}
}

func showSchematicCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show-schematic",
		Short: "Print cached schematic ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromConfig()
			if err != nil {
				return err
			}

			id, err := client.ShowSchematic()
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), id)
			return nil
		},
	}
}

func clearSchematicCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear-schematic",
		Short: "Remove cached schematic ID",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromConfig()
			if err != nil {
				return err
			}
			return client.ClearSchematic()
		},
	}
}
