package cilium

import (
	"errors"
	"os"

	"github.com/zerodi/cctl/internal/cilium"
	"github.com/zerodi/cctl/internal/configx"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func installCmd() *cobra.Command {
	defaultVersion := viper.GetString("cluster.ciliumVersion")
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install or upgrade Cilium via helm (kube-proxy-free, KubePrism ready)",
		RunE: func(cmd *cobra.Command, args []string) error {
			settings := configx.Cluster()
			version := viper.GetString("cluster.ciliumVersion")

			kcfg := ""
			if _, err := os.Stat(settings.KubeconfigPath); err == nil {
				kcfg = settings.KubeconfigPath
			} else if !errors.Is(err, os.ErrNotExist) {
				return err
			}

			log.Info().Str("version", version).Str("kubeconfig", kcfg).Msg("Installing Cilium")
			return cilium.Install(cmd.Context(), version, kcfg)
		},
	}

	cmd.Flags().String("version", defaultVersion, "Cilium version to install")
	_ = viper.BindPFlag("cluster.ciliumVersion", cmd.Flags().Lookup("version"))
	return cmd
}
