package cilium

import (
	"context"
	"fmt"
	"os"

	"github.com/zerodi/cctl/internal/executil"

	"github.com/rs/zerolog/log"
)

// Install installs or upgrades Cilium using helm, matching the legacy shell script.
func Install(ctx context.Context, version, kubeconfig string) error {
	if version == "" {
		return fmt.Errorf("cilium version is required")
	}
	if err := executil.EnsureCommands("helm", "kubectl"); err != nil {
		return err
	}

	env := withKubeconfig(kubeconfig)

	if err := executil.RunStreaming(ctx, env, "helm", "repo", "add", "cilium", "https://helm.cilium.io"); err != nil {
		log.Debug().Err(err).Msg("helm repo add cilium (non-fatal)")
	}
	if err := executil.RunStreaming(ctx, env, "helm", "repo", "update"); err != nil {
		log.Debug().Err(err).Msg("helm repo update (non-fatal)")
	}

	log.Info().Str("version", version).Msg("Installing Cilium via helm")
	args := []string{
		"upgrade", "--install", "cilium", "cilium/cilium",
		"--namespace", "kube-system",
		"--create-namespace",
		"--version", version,
		"--set", "kubeProxyReplacement=true",
		"--set", "k8sServiceHost=localhost",
		"--set", "k8sServicePort=7445",
		"--set", "routingMode=native",
		"--set", "ipam.mode=kubernetes",
		"--set", "hubble.enabled=true",
		"--set", "hubble.relay.enabled=true",
		"--set", "hubble.ui.enabled=true",
	}
	if err := executil.RunStreaming(ctx, env, "helm", args...); err != nil {
		return fmt.Errorf("helm upgrade --install cilium: %w", err)
	}

	log.Info().Msg("Waiting for Cilium DaemonSet rollout")
	if err := executil.RunStreaming(ctx, env, "kubectl", "-n", "kube-system", "rollout", "status", "ds/cilium", "--timeout=5m"); err != nil {
		return fmt.Errorf("kubectl rollout status cilium: %w", err)
	}

	if err := executil.RunStreaming(ctx, env, "kubectl", "-n", "kube-system", "get", "pods", "-l", "k8s-app=cilium", "-owide"); err != nil {
		log.Warn().Err(err).Msg("kubectl get pods (non-fatal)")
	}

	log.Info().Msg("Cilium installation complete")
	return nil
}

func withKubeconfig(path string) []string {
	if path == "" {
		return nil
	}
	env := os.Environ()
	return append(env, "KUBECONFIG="+path)
}
