package kind

import (
	"github.com/zerodi/cctl/internal/kindx"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func downCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "delete kind cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := viper.GetString("kind.name")
			if name == "" {
				name = "dev"
			}
			log.Info().Str("name", name).Msg("deleting kind cluster")
			return kindx.Delete(name)
		},
	}

	cmd.Flags().String("name", "dev", "kind cluster name")
	_ = viper.BindPFlag("kind.name", cmd.Flags().Lookup("name"))
	return cmd
}
