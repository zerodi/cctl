package kindx

import (
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"sigs.k8s.io/kind/pkg/cluster"
)

func Create(name, configPath string) error {
	provider := cluster.NewProvider()
	opts := []cluster.CreateOption{}
	if st, err := os.Stat(configPath); err == nil && !st.IsDir() {
		opts = append(opts, cluster.CreateWithConfigFile(configPath))
	}
	// Wait briefly for nodes to become ready
	opts = append(opts, cluster.CreateWithWaitForReady(2*time.Minute))

	log.Debug().Str("name", name).Str("config", configPath).Msg("kind: creating cluster")
	return provider.Create(name, opts...)
}

// Delete removes the kind cluster with the provided name.
func Delete(name string) error {
	provider := cluster.NewProvider()
	log.Debug().Str("name", name).Msg("kind: deleting cluster")
	return provider.Delete(name, "")
}

// Reset recreates the cluster by deleting then creating it again.
func Reset(name, configPath string) error {
	_ = Delete(name)
	return Create(name, configPath)
}
