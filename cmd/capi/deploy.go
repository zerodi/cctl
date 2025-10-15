/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package capi

import (
	"fmt"

	"github.com/zerodi/cctl/internal/capix"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func deployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "apply Cluster API manifests (kubectl apply)",
		RunE: func(cmd *cobra.Command, args []string) error {
			file := viper.GetString("capi.file")
			ns := viper.GetString("capi.namespace")
			if file == "" {
				return fmt.Errorf("manifest path is required: --file")
			}
			log.Info().Str("file", file).Str("namespace", ns).Msg("capi deploy")
			return capix.Apply(file, ns)
		},
	}

	cmd.Flags().StringP("file", "f", "configs/capi/templates/cluster.yaml", "file or directory with manifests")
	cmd.Flags().String("namespace", "default", "namespace for manifests")
	_ = viper.BindPFlag("capi.file", cmd.Flags().Lookup("file"))
	_ = viper.BindPFlag("capi.namespace", cmd.Flags().Lookup("namespace"))
	return cmd
}
