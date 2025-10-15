package proxmox

import (
	"time"

	"github.com/zerodi/cctl/internal/proxmox"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxmox",
		Short: "Proxmox helpers (Talos schematics, ISO uploads, VM templates)",
	}

	flags := cmd.PersistentFlags()
	flags.String("url", "", "Proxmox host/IP (without scheme)")
	flags.String("token-id", "", "Proxmox API token ID")
	flags.String("token-secret", "", "Proxmox API token secret")
	flags.String("node", "", "Proxmox node name")
	flags.String("iso-storage", "", "Proxmox storage target for ISO uploads (default: local)")
	flags.String("schematic-file", "", "Path to cached Talos schematic id")
	flags.String("schematic-yaml", "", "Talos factory schematic YAML input")
	flags.String("template-json", "", "Template JSON payload for VM creation")
	flags.Bool("skip-tls-verify", true, "Skip TLS verification for Proxmox API")
	flags.Duration("http-timeout", 60*time.Second, "HTTP timeout for Proxmox/Talos requests")

	_ = viper.BindPFlag("proxmox.url", flags.Lookup("url"))
	_ = viper.BindPFlag("proxmox.tokenID", flags.Lookup("token-id"))
	_ = viper.BindPFlag("proxmox.tokenSecret", flags.Lookup("token-secret"))
	_ = viper.BindPFlag("proxmox.node", flags.Lookup("node"))
	_ = viper.BindPFlag("proxmox.isoStorage", flags.Lookup("iso-storage"))
	_ = viper.BindPFlag("proxmox.schematicFile", flags.Lookup("schematic-file"))
	_ = viper.BindPFlag("proxmox.schematicYAML", flags.Lookup("schematic-yaml"))
	_ = viper.BindPFlag("proxmox.templateJSON", flags.Lookup("template-json"))
	_ = viper.BindPFlag("proxmox.skipTLSVerify", flags.Lookup("skip-tls-verify"))
	_ = viper.BindPFlag("proxmox.httpTimeout", flags.Lookup("http-timeout"))

	// Provide backwards compatibility with legacy environment variables.
	_ = viper.BindEnv("proxmox.url", "PROXMOX_URL")
	_ = viper.BindEnv("proxmox.tokenID", "PROXMOX_TOKEN")
	_ = viper.BindEnv("proxmox.tokenSecret", "PROXMOX_SECRET")
	_ = viper.BindEnv("proxmox.node", "PVE_NODE")
	_ = viper.BindEnv("proxmox.isoStorage", "PROXMOX_ISO_STORAGE")
	_ = viper.BindEnv("proxmox.schematicFile", "SCHEMATIC_FILE")
	_ = viper.BindEnv("proxmox.schematicYAML", "TALOS_SCHEMATIC_YAML")
	_ = viper.BindEnv("proxmox.templateJSON", "TEMPLATE_JSON")
	_ = viper.BindEnv("proxmox.skipTLSVerify", "PROXMOX_SKIP_TLS_VERIFY")

	cmd.AddCommand(refreshSchematicCmd())
	cmd.AddCommand(showSchematicCmd())
	cmd.AddCommand(clearSchematicCmd())
	cmd.AddCommand(getTalosImageCmd())
	cmd.AddCommand(createTemplateCmd())

	return cmd
}

func clientFromConfig() (*proxmox.Client, error) {
	timeout := viper.GetDuration("proxmox.httpTimeout")
	cfg := proxmox.Config{
		URL:                viper.GetString("proxmox.url"),
		TokenID:            viper.GetString("proxmox.tokenID"),
		Secret:             viper.GetString("proxmox.tokenSecret"),
		Node:               viper.GetString("proxmox.node"),
		ISOStorage:         viper.GetString("proxmox.isoStorage"),
		SchematicFile:      viper.GetString("proxmox.schematicFile"),
		TalosSchematicPath: viper.GetString("proxmox.schematicYAML"),
		TemplateJSONPath:   viper.GetString("proxmox.templateJSON"),
		SkipTLSVerify:      viper.GetBool("proxmox.skipTLSVerify"),
	}
	if timeout > 0 {
		cfg.Timeout = timeout
	}
	return proxmox.New(cfg)
}
