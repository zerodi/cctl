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

func getKubeconfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-kubeconfig",
		Short: "Wait for the Cluster API kubeconfig secret and write it to disk",
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

			pattern := fmt.Sprintf("%s-kubeconfig", settings.Name)
			log.Info().Str("secretPattern", pattern).Str("namespace", settings.Namespace).Msg("Waiting for kubeconfig secret")

			name, err := client.WaitForSecret(cmd.Context(), pattern, settings.Name, timeout)
			if err != nil {
				return err
			}
			log.Info().Str("secret", name).Msg("Found kubeconfig secret")

			if err := client.ExtractSecretToFile(cmd.Context(), name, "kubeconfig", settings.KubeconfigPath); err != nil {
				return err
			}
			log.Info().Str("path", settings.KubeconfigPath).Msg("Wrote kubeconfig")
			return nil
		},
	}

	cmd.Flags().Duration("timeout", 20*time.Minute, "Maximum time to wait for the secret")
	_ = viper.BindPFlag("secrets.timeout", cmd.Flags().Lookup("timeout"))
	return cmd
}

func existingPath(path string) string {
	if path == "" {
		return ""
	}
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}
