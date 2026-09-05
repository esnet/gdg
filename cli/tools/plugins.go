package tools

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/domain"
	"github.com/esnet/gdg/internal/adapter/plugins/lookup"
	"github.com/esnet/gdg/internal/adapter/plugins/registry"
	gdgconfig "github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/ports/outbound"
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
			newPluginsRekeyCmd(),
			newPluginsCipherCmd(),
			newPluginsLookupCmd(),
		},
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Aliases = []string{"plugin"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, r *domain.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
		Short: "Browse and manage cipher and lookup plugins",
		Long:  "Browse available cipher and lookup plugins from the registry, and manage plugin configuration.",
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
			cmd.Flags().String("type", "cipher", "Plugin type to list: cipher, lookup, or all")
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, r *domain.RootCommand, args []string) error {
			client := buildRegistryClient(cd, r)

			pluginType, _ := cd.CobraCommand.Flags().GetString("type")

			fetch := client.CipherPlugins
			switch strings.ToLower(pluginType) {
			case "", "cipher":
				fetch = client.CipherPlugins
			case "lookup":
				fetch = client.LookupPlugins
			case "all":
				fetch = client.All
			default:
				return fmt.Errorf("unknown plugin type %q: must be one of cipher, lookup, all", pluginType)
			}

			plugins, err := fetch()
			if err != nil {
				return err
			}

			r.TableObj.AppendHeader(table.Row{"Name", "Type", "Version", "GDG Restrictions", "Config Fields", "Description"})
			for _, p := range plugins {
				for _, v := range p.Versions {
					var restrictions string
					restrictions = fmt.Sprintf(
						"gdg_min_ver: %s\nvalid: %v",
						v.MinimumVersion, v.IsValid())

					if v.MaximumVersion != "" {
						restrictions = fmt.Sprintf("%s\n%v", restrictions, v.MaximumVersion)
					}

					r.TableObj.AppendRow(table.Row{
						p.Name,
						p.Type,
						v.Version,
						restrictions,
						strings.Join(v.ConfigFields, "\n"),
						p.Description,
					})
				}
			}

			r.Render(cd.CobraCommand, plugins)
			return nil
		},
		Short: "List available cipher or lookup plugins",
		Long: `List plugins available in the GDG plugin registry.

For each plugin its name, available versions, required configuration fields,
and a short description are displayed.

Use --type to select which plugin family to list: "cipher" (the default,
for backward compatibility), "lookup", or "all" for every registry entry
regardless of type.

The registry is loaded from a local file (--registry-file) when provided,
otherwise fetched from the configured URL (--registry-url, or the value of
global.plugin_registry_url in gdg.yml, or the built-in default).`,
	}
}

// newPluginsRekeyCmd returns the "gdg tools plugins rekey" command.
// It launches an interactive TUI that walks the user through re-encrypting all
// on-disk GDG files after switching cipher plugins or disabling encryption.
// No live Grafana connection is required (registered in noLoginGroups via "plugins").
func newPluginsRekeyCmd() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "rekey",
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Flags().String("registry-url", "", "Override the remote registry URL")
			cmd.Flags().String("registry-file", "", "Load registry from a local file instead of fetching remotely")
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, r *domain.RootCommand, args []string) error {
			regClient := buildRegistryClient(cd, r)
			return gdgconfig.RunRekey(r.ConfigSvc(), regClient)
		},
		Short: "Re-encrypt GDG files after changing cipher plugin",
		Long: `Interactively re-encrypt all on-disk GDG files when switching to a
different cipher plugin or disabling encryption entirely.

The TUI will:
  1. Show the currently active cipher plugin configuration.
  2. Let you choose to switch to a new plugin, disable encryption, or cancel.
  3. Scan all affected files for the selected context and verify decodability.
  4. Let you select which files to include (deselect to skip individual files).
  5. Run the migration with optional backup, then update gdg.yml.

A backup of all modified files is created by default before any changes are made.`,
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

// newPluginsCipherCmd returns the "gdg tools plugins cipher" command.
// Moved here from "gdg tools helpers cipher" so all cipher-plugin
// interaction (browsing the registry via list, migrating via rekey, and
// now encoding/decoding values directly) lives under one "plugins"
// command tree instead of being split across "helpers" and "plugins".
func newPluginsCipherCmd() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "cipher",
		Short: "Cipher Helpers",
		Long:  "Cipher Helpers",
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Aliases = []string{"c", "ciphers"}
			cmd.PersistentFlags().StringP("file", "f", "", "file to encode/decode")
			cmd.PersistentFlags().StringP("value", "", "", "value to encode/decode")
		},
		CommandsList: []simplecobra.Commander{
			newPluginsCipherEncodeCmd(),
			newPluginsCipherDecodeCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

// newPluginsLookupCmd returns the "gdg tools plugins lookup" command.
// It groups lookup-plugin diagnostics; currently just "test", a dry-run
// resolver invocation that requires no live Grafana connection.
func newPluginsLookupCmd() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "lookup",
		Short: "Lookup plugin helpers",
		Long:  "Inspect and test lookup plugins (e.g. Google Secret Manager) configured in gdg.yml.",
		CommandsList: []simplecobra.Commander{
			newPluginsLookupTestCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, r *domain.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

// newPluginsLookupTestCmd returns the "gdg tools plugins lookup test" command.
// Usage: gdg tools plugins lookup test <provider> <key>
//
// It builds a lookup.Resolver from the currently loaded configuration
// (the same plugins.lookup config block used at runtime) and resolves
// "lookup:<provider>:<key>" against it, printing the result or a
// descriptive error. No live Grafana connection is required, which lets
// users verify lookup plugin configuration (e.g. GSM credentials) on its
// own before running a full gdg command against Grafana.
func newPluginsLookupTestCmd() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "test",
		Short: "Resolve a lookup reference using the configured plugins",
		Long: `Resolve "lookup:<provider>:<key>[.<json_field>]" against the lookup
plugins configured under plugins.lookup in gdg.yml.

Example:

  gdg tools plugins lookup test gsm projects/my-gcp-project/secrets/grafana-api-token/versions/latest.token

This performs the same resolution GDG performs internally against the
Grafana Token/Password fields, without needing a live Grafana connection,
so lookup plugin configuration can be verified in isolation.`,
		// WithCFunc runs during command-tree Init(), before any execution, so
		// cmd.Args set here is visible to Cobra's ValidateArgs call (see
		// newContextCopy in context.go for the same pattern).
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Args = func(cmd *cobra.Command, args []string) error {
				if len(args) < 2 {
					return errors.New("requires a provider and key argument")
				}
				return nil
			}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, r *domain.RootCommand, args []string) error {
			// Guard against the validator not firing (defensive belt-and-suspenders).
			if len(args) < 2 {
				return errors.New("requires a provider and key argument")
			}
			provider := args[0]
			key := args[1]

			cfg := r.ConfigSvc()
			resolver, err := lookup.NewResolver(&cfg.PluginConfig)
			if err != nil {
				return fmt.Errorf("initializing lookup plugins: %w", err)
			}

			ref := fmt.Sprintf("%s%s:%s", outbound.LookupPrefix, provider, key)
			value, err := resolver.Resolve(ref)
			if err != nil {
				return fmt.Errorf("resolving %q: %w", ref, err)
			}

			fmt.Println(value)
			return nil
		},
	}
}

// newPluginsCipherEncodeCmd returns the "gdg tools plugins cipher encode" command.
func newPluginsCipherEncodeCmd() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "encode",
		Short: "apply cipher to string",
		Long:  "apply cipher to string",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			fileName, _ := cd.CobraCommand.Flags().GetString("file")
			value, _ := cd.CobraCommand.Flags().GetString("value")
			if fileName != "" && value != "" {
				log.Fatal("either a value or a file must be specified, not both")
			}
			if value != "" {
				result := rootCmd.GrafanaSvc().EncodeValue(value)
				slog.Info("Encoded result:")
				fmt.Println(result)
			} else {
				data, err := os.ReadFile(fileName) // #nosec G304
				if err != nil {
					log.Fatal("Error reading file", "file", fileName, "err", err)
				}

				result := rootCmd.GrafanaSvc().EncodeValue(string(data))
				if result != "" {
					err = os.WriteFile(fileName, []byte(result), 0o600) // #nosec G703 TODO:revisit
					if err != nil {
						log.Fatal("Error writing file", "file", fileName, "err", err)
					} else {
						slog.Info("File has been encrypted", "file", fileName)
					}
				}
			}

			return nil
		},
	}
}

// newPluginsCipherDecodeCmd returns the "gdg tools plugins cipher decode" command.
func newPluginsCipherDecodeCmd() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "decode",
		Short: "decode string using cipher plugin",
		Long:  "decode string using cipher plugin",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			fileName, _ := cd.CobraCommand.Flags().GetString("file")
			value, _ := cd.CobraCommand.Flags().GetString("value")
			if fileName != "" && value != "" {
				log.Fatal("either a value or a file must be specified, not both")
			}
			if value != "" {
				result := rootCmd.GrafanaSvc().DecodeValue(value)
				slog.Info("Decoded result")
				fmt.Println(result)
			} else {
				data, err := os.ReadFile(fileName) // #nosec G304
				if err != nil {
					log.Fatal("Error reading file", "file", fileName, "err", err)
				}

				result := rootCmd.GrafanaSvc().DecodeValue(string(data))
				if result != "" {
					err = os.WriteFile(fileName, []byte(result), 0o600) // #nosec G703 TODO:revisit
					if err != nil {
						log.Fatal("Error writing file", "file", fileName, "err", err)
					} else {
						slog.Info("File has been decrypted", "file", fileName)
					}
				}
			}
			return nil
		},
	}
}
