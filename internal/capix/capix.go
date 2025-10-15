package capix

import (
	"context"
	"os"
	"os/exec"

	"github.com/rs/zerolog/log"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client"
)

func Init(clusterConfig, core string, boots, cps, infras []string, kubeconfig string) error {
	c, err := client.New(context.Background(), clusterConfig)
	if err != nil {
		return err
	}

	opts := client.InitOptions{
		CoreProvider:            core,
		BootstrapProviders:      boots,
		ControlPlaneProviders:   cps,
		InfrastructureProviders: infras,
	}
	if kubeconfig != "" {
		opts.Kubeconfig = client.Kubeconfig{Path: kubeconfig}
	}

	log.Debug().
		Str("core", core).
		Strs("bootstrap", boots).
		Strs("controlPlane", cps).
		Strs("infrastructure", infras).
		Msg("clusterctl: init")

	_, err = c.Init(context.Background(), opts)
	return err

}

// Apply runs `kubectl apply -f` on the provided file or directory.
func Apply(path, namespace string) error {
	args := []string{"apply", "-f", path}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	cmd := exec.Command("kubectl", args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}
