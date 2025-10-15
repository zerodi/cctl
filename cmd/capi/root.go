/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package capi

import (
	"github.com/spf13/cobra"
)

// New returns the `capi` command group with all subcommands.
func New() *cobra.Command {
	cmd := &cobra.Command{Use: "capi", Short: "Commands for Cluster API"}
	cmd.AddCommand(initCmd())
	cmd.AddCommand(deployCmd())
	return cmd
}
