package secrets

import (
	"fmt"
	"os"
	"time"

	"github.com/zerodi/cctl/internal/configx"
	"github.com/zerodi/cctl/internal/executil"
	"github.com/zerodi/cctl/internal/kubex"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func getTalosconfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-talosconfig",
		Short: "Wait for the Cluster API talosconfig secret and write it to disk",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := executil.EnsureCommands("kubectl"); err != nil {
				return err
			}

			settings := configx.Cluster()
			if err := os.MkdirAll(settings.OutDir, 0o755); err != nil {
				return fmt.Errorf("ensure output dir: %w", err)
			}

			kcfg := existingPath(settings.KubeconfigPath)
			client := kubex.New(kcfg, settings.Namespace)
			timeout := viper.GetDuration("secrets.timeout")
			if timeout == 0 {
				timeout = 20 * time.Minute
			}

			pattern := fmt.Sprintf("%s-talosconfig", settings.Name)
			log.Info().Str("secretPattern", pattern).Str("namespace", settings.Namespace).Msg("Waiting for talosconfig secret")

			name, err := client.WaitForSecret(cmd.Context(), pattern, settings.Name, timeout)
			if err != nil {
				return err
			}
			log.Info().Str("secret", name).Msg("Found talosconfig secret")

			if err := client.ExtractSecretToFile(cmd.Context(), name, "talosconfig", settings.TalosconfigPath); err != nil {
				return err
			}
			log.Info().Str("path", settings.TalosconfigPath).Msg("Wrote talosconfig")
			return nil
		},
	}

	cmd.Flags().Duration("timeout", 20*time.Minute, "Maximum time to wait for the secret")
	_ = viper.BindPFlag("secrets.timeout", cmd.Flags().Lookup("timeout"))
	return cmd
}
