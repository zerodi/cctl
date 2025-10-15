package kind

import (
	"github.com/zerodi/cctl/internal/kindx"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func resetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "recreate kind cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			name := viper.GetString("kind.name")
			cfg := viper.GetString("kind.config")
			if name == "" {
				name = "dev"
			}
			log.Info().Str("name", name).Str("config", cfg).Msg("reset kind cluster")
			return kindx.Reset(name, cfg)
		},
	}

	cmd.Flags().String("name", "dev", "kind cluster name")
	cmd.Flags().String("config", "configs/kind.yaml", "path to kind config")
	_ = viper.BindPFlag("kind.name", cmd.Flags().Lookup("name"))
	_ = viper.BindPFlag("kind.config", cmd.Flags().Lookup("config"))
	return cmd
}
