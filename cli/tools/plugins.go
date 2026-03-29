package tools

import (
	"context"
	"strings"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/domain"
	"github.com/esnet/gdg/internal/adapter/plugins/registry"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

// newPluginsCmd returns the "gdg tools plugins" parent command.
// All subcommands operate on local config and the remote plugin registry;
// none require a live Grafana connection (registered in noLoginGroups).
func newPluginsCmd() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "plugins",
		CommandsList: []simplecobra.Commander{
			newPluginsListCmd(),
		},
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Aliases = []string{"plugin"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, r *domain.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
		Short: "Browse and manage cipher plugins",
		Long:  "Browse available cipher plugins from the registry and manage plugin configuration.",
	}
}

// newPluginsListCmd returns the "gdg tools plugins list" command.
// It fetches the plugin registry (from a local file or remote URL) and
// renders a table of all available cipher plugins, their versions, and
// required configuration fields.
func newPluginsListCmd() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "list",
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Flags().String("registry-url", "", "Override the remote registry URL")
			cmd.Flags().String("registry-file", "", "Load registry from a local file instead of fetching remotely")
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, r *domain.RootCommand, args []string) error {
			client := buildRegistryClient(cd, r)

			plugins, err := client.CipherPlugins()
			if err != nil {
				return err
			}

			r.TableObj.AppendHeader(table.Row{"Name", "Type", "Version", "Config Fields", "Description"})
			for _, p := range plugins {
				for _, v := range p.Versions {
					r.TableObj.AppendRow(table.Row{
						p.Name,
						p.Type,
						v.Version,
						strings.Join(v.ConfigFields, ", "),
						p.Description,
					})
				}
			}

			r.Render(cd.CobraCommand, plugins)
			return nil
		},
		Short: "List available cipher plugins",
		Long: `List all cipher plugins available in the GDG plugin registry.

For each plugin its name, available versions, required configuration fields,
and a short description are displayed.

The registry is loaded from a local file (--registry-file) when provided,
otherwise fetched from the configured URL (--registry-url, or the value of
global.plugin_registry_url in gdg.yml, or the built-in default).`,
	}
}

// buildRegistryClient constructs a registry.Client from CLI flags and config,
// applying the precedence: --registry-file > --registry-url > config file values.
func buildRegistryClient(cd *simplecobra.Commandeer, r *domain.RootCommand) *registry.Client {
	flagFile, _ := cd.CobraCommand.Flags().GetString("registry-file")
	flagURL, _ := cd.CobraCommand.Flags().GetString("registry-url")

	cfg := registry.ClientConfig{}

	switch {
	case flagFile != "":
		cfg.FilePath = flagFile
	case flagURL != "":
		cfg.URL = flagURL
	default:
		globals := r.ConfigSvc().GetAppGlobals()
		cfg.FilePath = globals.PluginRegistryFile
		cfg.URL = globals.PluginRegistryURL
		// If both are still empty, registry.Client defaults to domain.RegistryDefaultURL.
	}

	return registry.NewClient(cfg)
}
