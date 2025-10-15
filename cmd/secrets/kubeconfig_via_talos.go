package secrets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zerodi/cctl/internal/configx"
	"github.com/zerodi/cctl/internal/executil"
	"github.com/zerodi/cctl/internal/kubex"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func kubeconfigViaTalosCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubeconfig-via-talos",
		Short: "Generate a kubeconfig using talosctl and the Talos control plane",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := executil.EnsureCommands("kubectl", "talosctl"); err != nil {
				return err
			}

			settings := configx.Cluster()
			if err := os.MkdirAll(settings.OutDir, 0o755); err != nil {
				return fmt.Errorf("ensure output dir: %w", err)
			}

			if _, err := os.Stat(settings.TalosconfigPath); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("talosconfig not found at %s (run: secrets get-talosconfig)", settings.TalosconfigPath)
				}
				return err
			}

			if _, err := os.Stat(settings.KubeconfigPath); err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("kubeconfig not found at %s (run: secrets get-kubeconfig)", settings.KubeconfigPath)
				}
				return err
			}

			client := kubex.New(settings.KubeconfigPath, settings.Namespace)
			ctx := cmd.Context()

			ips, err := client.DiscoverControlPlaneIPs(ctx)
			if err != nil {
				return err
			}
			if len(ips) == 0 {
				return fmt.Errorf("no control-plane nodes discovered yet")
			}
			log.Info().Strs("ips", ips).Msg("Talos control plane endpoints")

			if err := executil.RunStreaming(ctx, nil, "talosctl", append([]string{"--talosconfig", settings.TalosconfigPath, "config", "endpoints"}, ips...)...); err != nil {
				return fmt.Errorf("talosctl config endpoints: %w", err)
			}
			if err := executil.RunStreaming(ctx, nil, "talosctl", append([]string{"--talosconfig", settings.TalosconfigPath, "config", "nodes"}, ips...)...); err != nil {
				return fmt.Errorf("talosctl config nodes: %w", err)
			}

			cp0 := ips[0]
			target := filepath.Join(settings.OutDir, fmt.Sprintf("kubeconfig-%s-talosctl", settings.Name))

			talosArgs := []string{
				"--talosconfig", settings.TalosconfigPath,
				"kubeconfig",
				"--nodes", cp0,
				"--endpoints", fmt.Sprintf("%s:7445", cp0),
				"--force",
			}

			stdout, stderr, err := executil.RunCapture(ctx, nil, "talosctl", talosArgs...)
			if err != nil {
				return fmt.Errorf("talosctl kubeconfig: %v: %s", err, strings.TrimSpace(stderr))
			}

			if err := os.WriteFile(target, []byte(stdout), 0o600); err != nil {
				return fmt.Errorf("write %s: %w", target, err)
			}

			log.Info().Str("path", target).Msg("Saved talosctl-generated kubeconfig")
			return nil
		},
	}

	return cmd
}
