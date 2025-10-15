package configx

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	defaultClusterName   = "coffee-cluster"
	defaultNamespace     = "default"
	defaultOutDir        = "./out"
	defaultCiliumVersion = "1.16.4"
)

func init() {
	viper.SetDefault("cluster.name", defaultClusterName)
	viper.SetDefault("cluster.namespace", defaultNamespace)
	viper.SetDefault("cluster.outDir", defaultOutDir)
	viper.SetDefault("cluster.ciliumVersion", defaultCiliumVersion)
}

// ClusterSettings captures reusable Cluster API defaults.
type ClusterSettings struct {
	Name            string
	Namespace       string
	OutDir          string
	KubeconfigPath  string
	TalosconfigPath string
	CiliumVersion   string
}

// Cluster loads settings from Viper, applying script-compatible fallbacks.
func Cluster() ClusterSettings {
	name := viper.GetString("cluster.name")
	if name == "" {
		name = defaultClusterName
	}

	ns := viper.GetString("cluster.namespace")
	if ns == "" {
		ns = defaultNamespace
	}

	out := viper.GetString("cluster.outDir")
	if out == "" {
		out = defaultOutDir
	}

	kubeconfig := viper.GetString("cluster.kubeconfigPath")
	if kubeconfig == "" {
		kubeconfig = filepath.Join(out, fmt.Sprintf("kubeconfig-%s", name))
	}

	talosconfig := viper.GetString("cluster.talosconfigPath")
	if talosconfig == "" {
		talosconfig = filepath.Join(out, fmt.Sprintf("talosconfig-%s", name))
	}

	ciliumVersion := viper.GetString("cluster.ciliumVersion")
	if ciliumVersion == "" {
		ciliumVersion = defaultCiliumVersion
	}

	return ClusterSettings{
		Name:            name,
		Namespace:       ns,
		OutDir:          out,
		KubeconfigPath:  kubeconfig,
		TalosconfigPath: talosconfig,
		CiliumVersion:   ciliumVersion,
	}
}
