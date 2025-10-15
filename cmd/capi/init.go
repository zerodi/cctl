/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package capi

import (
	"github.com/zerodi/cctl/internal/capix"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "initialize Cluster API providers via clusterctl",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := viper.GetString("capi.clusterctl_config")
			core := viper.GetString("capi.core")
			boots := viper.GetStringSlice("capi.bootstrap")
			cps := viper.GetStringSlice("capi.controlplane")
			infras := viper.GetStringSlice("capi.infrastructure")
			kubeconf := viper.GetString("capi.kubeconfig")

			log.Info().
				Str("clusterctl", cfg).
				Str("core", core).
				Strs("bootstrap", boots).
				Strs("controlPlane", cps).
				Strs("infrastructure", infras).
				Msg("capi init")

			return capix.Init(cfg, core, boots, cps, infras, kubeconf)
		},
	}

	cmd.Flags().String("clusterctl-config", "configs/capi/clusterctl.yaml", "path to clusterctl.yaml")
	cmd.Flags().String("core", "cluster-api", "CoreProvider (usually 'cluster-api')")
	cmd.Flags().StringSlice("bootstrap", []string{"kubeadm"}, "BootstrapProviders (comma separated)")
	cmd.Flags().StringSlice("control-plane", []string{"kubeadm"}, "ControlPlaneProviders (comma separated)")
	cmd.Flags().StringSlice("infrastructure", []string{"docker"}, "InfrastructureProviders (comma separated)")
	cmd.Flags().String("kubeconfig", "", "path to target cluster kubeconfig (optional)")

	_ = viper.BindPFlag("capi.clusterctl_config", cmd.Flags().Lookup("clusterctl-config"))
	_ = viper.BindPFlag("capi.core", cmd.Flags().Lookup("core"))
	_ = viper.BindPFlag("capi.bootstrap", cmd.Flags().Lookup("bootstrap"))
	_ = viper.BindPFlag("capi.controlplane", cmd.Flags().Lookup("control-plane"))
	_ = viper.BindPFlag("capi.infrastructure", cmd.Flags().Lookup("infrastructure"))
	_ = viper.BindPFlag("capi.kubeconfig", cmd.Flags().Lookup("kubeconfig"))
	return cmd
}
