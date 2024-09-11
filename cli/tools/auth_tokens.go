package tools

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"slices"
	"sort"
	"strconv"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func newTokensCmd() simplecobra.Commander {
	description := "Provides some utility to help the user manage their API token keys"
	return &support.SimpleCommand{
		NameP: "tokens",
		Short: description,
		Long:  description,
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"token", "apikeys"}
		},
		CommandsList: []simplecobra.Commander{
			newListTokensCmd(),
			newDeleteTokenCmd(),
			newNewTokenCmd(),
		},
	}
}

func newListTokensCmd() simplecobra.Commander {
	description := "List API Keys"
	return &support.SimpleCommand{
		NameP:        "list",
		Short:        description,
		Long:         description,
		CommandsList: []simplecobra.Commander{},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"id", "name", "role", "expiration"})
			apiKeys := rootCmd.GrafanaSvc().ListAPIKeys()
			sort.SliceStable(apiKeys, func(i, j int) bool {
				return apiKeys[i].ID < apiKeys[j].ID
			})
			if len(apiKeys) == 0 {
				slog.Info("No apiKeys found")
			} else {
				for _, apiKey := range apiKeys {
					var formattedDate string = apiKey.Expiration.String()
					date, _ := apiKey.Expiration.Value()
					if date.(string) == "0001-01-01T00:00:00.000Z" {
						formattedDate = "No Expiration"
					}

					rootCmd.TableObj.AppendRow(table.Row{apiKey.ID, apiKey.Name, apiKey.Role, formattedDate})
				}
				rootCmd.Render(cd.CobraCommand, apiKeys)
			}
			return nil
		},
	}
}

func newDeleteTokenCmd() simplecobra.Commander {
	description := "delete all Tokens from grafana"
	return &support.SimpleCommand{
		NameP:        "clear",
		Short:        description,
		Long:         description,
		CommandsList: []simplecobra.Commander{},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			savedFiles := rootCmd.GrafanaSvc().DeleteAllTokens()
			slog.Info("Delete Tokens for context: ", "context", config.Config().GetGDGConfig().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"type", "filename"})
			if len(savedFiles) == 0 {
				slog.Info("No Tokens found")
			} else {
				for _, file := range savedFiles {
					rootCmd.TableObj.AppendRow(table.Row{"user", file})
				}
				rootCmd.Render(cd.CobraCommand, savedFiles)
			}
			return nil
		},
	}
}

func newNewTokenCmd() simplecobra.Commander {
	description := "new <name> <role> [ttl in seconds]"
	return &support.SimpleCommand{
		NameP:        "new",
		Short:        description,
		Long:         description,
		CommandsList: []simplecobra.Commander{},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			if len(args) < 2 {
				return errors.New("requires a key name and a role('admin','viewer','editor') [ttl optional] ")
			}
			name := args[0]
			role := args[1]
			ttl := "0"
			if len(args) > 2 {
				ttl = args[2]
			}
			var (
				expiration int64
				err        error
			)

			expiration, err = strconv.ParseInt(ttl, 10, 64)
			if err != nil {
				expiration = 0
			}

			if !slices.Contains([]string{"admin", "editor", "viewer"}, role) {
				log.Fatal("Invalid role specified")
			}
			key, err := rootCmd.GrafanaSvc().CreateAPIKey(name, role, expiration)
			if err != nil {
				log.Fatal("unable to create api key", "err", err)
			} else {

				rootCmd.TableObj.AppendHeader(table.Row{"id", "name", "token"})
				rootCmd.TableObj.AppendRow(table.Row{key.ID, key.Name, key.Key})
				rootCmd.Render(cd.CobraCommand, map[string]interface{}{"id": key.ID, "name": key.Name, "token": key.Key})
			}

			return nil
		},
	}
}
