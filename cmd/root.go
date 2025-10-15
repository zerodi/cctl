/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/zerodi/cctl/cmd/capi"
	"github.com/zerodi/cctl/cmd/cilium"
	"github.com/zerodi/cctl/cmd/kind"
	"github.com/zerodi/cctl/cmd/proxmox"
	"github.com/zerodi/cctl/cmd/secrets"
	"github.com/zerodi/cctl/internal/configx"
	logx "github.com/zerodi/cctl/internal/logx"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	version = "0.1.0"
	cfgFile string
)

func Execute() error { return rootCmd().Execute() }

func rootCmd() *cobra.Command {
	bindLegacyEnv()
	clusterDefaults := configx.Cluster()

	root := &cobra.Command{
		Use:   "cctl",
		Short: "CLI Tool for managing Kind and Cluster API",
		// Uncomment the following line if your bare application
		// has an action associated with it:
		// Run: func(cmd *cobra.Command, args []string) { },
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			viper.SetEnvPrefix("CCTL")
			viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
			viper.AutomaticEnv()

			// Config file (if provided)
			if cfgFile != "" {
				viper.SetConfigFile(cfgFile)
				if err := viper.ReadInConfig(); err != nil {
					return fmt.Errorf("read config: %w", err)
				}
			}

			// Logger
			logx.Configure(viper.GetString("log.format"), viper.GetBool("debug"))
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) { _ = cmd.Help() },
	}

	// Global flags
	root.PersistentFlags().Bool("debug", false, "enable verbose logging")
	_ = viper.BindPFlag("debug", root.PersistentFlags().Lookup("debug"))

	root.PersistentFlags().StringVar(&cfgFile, "config", "", "path to config file (yaml|json|toml)")
	root.PersistentFlags().String("log-format", "console", "log format: console|json")
	_ = viper.BindPFlag("log.format", root.PersistentFlags().Lookup("log-format"))

	root.PersistentFlags().String("cluster-name", clusterDefaults.Name, "Cluster API workload cluster name")
	_ = viper.BindPFlag("cluster.name", root.PersistentFlags().Lookup("cluster-name"))

	root.PersistentFlags().String("namespace", clusterDefaults.Namespace, "Namespace for Cluster API objects")
	_ = viper.BindPFlag("cluster.namespace", root.PersistentFlags().Lookup("namespace"))

	root.PersistentFlags().String("out-dir", clusterDefaults.OutDir, "Output directory for generated artifacts")
	_ = viper.BindPFlag("cluster.outDir", root.PersistentFlags().Lookup("out-dir"))

	root.PersistentFlags().String("kubeconfig-path", clusterDefaults.KubeconfigPath, "Path to workload cluster kubeconfig")
	_ = viper.BindPFlag("cluster.kubeconfigPath", root.PersistentFlags().Lookup("kubeconfig-path"))

	root.PersistentFlags().String("talosconfig-path", clusterDefaults.TalosconfigPath, "Path to workload cluster talosconfig")
	_ = viper.BindPFlag("cluster.talosconfigPath", root.PersistentFlags().Lookup("talosconfig-path"))

	// Command groups
	root.AddCommand(kind.New())
	root.AddCommand(capi.New())
	root.AddCommand(proxmox.New())
	root.AddCommand(cilium.New())
	root.AddCommand(secrets.New())

	// Version
	root.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "show version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	})

	return root
}

func bindLegacyEnv() {
	_ = viper.BindEnv("cluster.name", "CLUSTER")
	_ = viper.BindEnv("cluster.namespace", "NS")
	_ = viper.BindEnv("cluster.outDir", "OUT_DIR")
	_ = viper.BindEnv("cluster.kubeconfigPath", "KUBECONFIG_PATH")
	_ = viper.BindEnv("cluster.talosconfigPath", "TALOSCONFIG_PATH")
	_ = viper.BindEnv("cluster.ciliumVersion", "CILIUM_VER")
}
