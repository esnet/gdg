package tools

import (
	"context"
	"errors"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/jedib0t/go-pretty/v6/table"
	"log"
	"log/slog"

	"github.com/spf13/cobra"
	"slices"
	"sort"
	"strconv"
)

func newServiceAccountCmd() simplecobra.Commander {
	description := "Manage api service-account"
	return &support.SimpleCommand{
		NameP: "service-accounts",
		Short: description,
		Long:  description,
		CommandsList: []simplecobra.Commander{
			newListServiceAccountCmd(),
			newDeleteServiceAccountCmd(),
			newDeleteServiceAccountTokensCmd(),
			newServiceAccount(),
			newServiceAccountTokenCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"service-account", "svcAcct", "svcAccts", "svc"}
		},
	}
}

func newListServiceAccountCmd() simplecobra.Commander {
	description := "List Service Accounts"
	return &support.SimpleCommand{
		NameP:        "list",
		Short:        description,
		Long:         description,
		CommandsList: []simplecobra.Commander{},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"id", "service name", "role", "tokens", "token id", "token name", "expiration"})
			apiKeys := rootCmd.GrafanaSvc().ListServiceAccounts()
			sort.SliceStable(apiKeys, func(i, j int) bool {
				return apiKeys[i].ServiceAccount.ID < apiKeys[j].ServiceAccount.ID
			})
			if len(apiKeys) == 0 {
				slog.Info("No Service Accounts found")
			} else {
				for _, apiKey := range apiKeys {

					rootCmd.TableObj.AppendRow(table.Row{apiKey.ServiceAccount.ID, apiKey.ServiceAccount.Name, apiKey.ServiceAccount.Role, apiKey.ServiceAccount.Tokens})
					if apiKey.Tokens != nil {
						sort.SliceStable(apiKey.Tokens, func(i, j int) bool {
							return apiKey.Tokens[i].ID < apiKey.Tokens[j].ID
						})
						for _, token := range apiKey.Tokens {
							var formattedDate string = token.Expiration.String()
							date, _ := token.Expiration.Value()
							if date.(string) == "0001-01-01T00:00:00.000Z" {
								formattedDate = "No Expiration"
							}
							rootCmd.TableObj.AppendRow(table.Row{"", "", "", "", token.ID, token.Name, formattedDate})
						}
					}
				}
				rootCmd.TableObj.Render()
			}

			return nil
		},
	}
}

func newDeleteServiceAccountTokensCmd() simplecobra.Commander {
	description := "delete all tokens for Service Account from grafana"
	return &support.SimpleCommand{
		NameP:        "clearTokens",
		Short:        description,
		Long:         description,
		CommandsList: []simplecobra.Commander{},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a service account ID to be specified")
			}
			idStr := args[0]
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				log.Fatalf("unable to parse %s as a valid numeric value", idStr)
			}

			slog.Info("Deleting Service Accounts Tokens for context",
				"serviceAccountId", id,
				"context", config.Config().AppConfig.GetContext())
			savedFiles := rootCmd.GrafanaSvc().DeleteServiceAccountTokens(id)
			rootCmd.TableObj.AppendHeader(table.Row{"serviceID", "type", "token_name"})
			if len(savedFiles) == 0 {
				slog.Info("No Service Accounts tokens found")
			} else {
				for _, token := range savedFiles {
					rootCmd.TableObj.AppendRow(table.Row{id, "service token", token})
				}
				rootCmd.TableObj.Render()
			}
			return nil
		},
	}
}

func newDeleteServiceAccountCmd() simplecobra.Commander {
	description := "delete all Service Accounts from grafana"
	return &support.SimpleCommand{
		NameP:        "clear",
		Short:        description,
		Long:         description,
		CommandsList: []simplecobra.Commander{},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			savedFiles := rootCmd.GrafanaSvc().DeleteAllServiceAccounts()
			slog.Info("Delete Service Accounts for context", "context", config.Config().AppConfig.GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"type", "filename"})
			if len(savedFiles) == 0 {
				slog.Info("No Service Accounts found")
			} else {
				for _, file := range savedFiles {
					rootCmd.TableObj.AppendRow(table.Row{"user", file})
				}
				rootCmd.TableObj.Render()
			}
			return nil
		},
	}
}

func newServiceAccount() simplecobra.Commander {
	description := "newService <serviceName> <role> [ttl in seconds]"
	return &support.SimpleCommand{
		NameP:        "newService",
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
			serviceAcct, err := rootCmd.GrafanaSvc().CreateServiceAccount(name, role, expiration)
			if err != nil {
				log.Fatal("unable to create api key", "error", err)
			} else {

				rootCmd.TableObj.AppendHeader(table.Row{"id", "name", "role"})
				rootCmd.TableObj.AppendRow(table.Row{serviceAcct.ID, serviceAcct.Name, serviceAcct.Role})
				rootCmd.TableObj.Render()
			}
			return nil
		},
	}
}

func newServiceAccountTokenCmd() simplecobra.Commander {
	description := "newToken <serviceAccountID> <name> [ttl in seconds]"
	return &support.SimpleCommand{
		NameP:        "newToken",
		Short:        description,
		Long:         description,
		CommandsList: []simplecobra.Commander{newTokensCmd()},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			if len(args) < 2 {
				return errors.New("requires a service-account ID and token name [ttl optional] ")
			}
			serviceIDRaw := args[0]
			name := args[1]
			ttl := "0"
			if len(args) > 2 {
				ttl = args[2]
			}
			var (
				expiration int64
				err        error
			)

			serviceID, err := strconv.ParseInt(serviceIDRaw, 10, 64)
			if err != nil {
				log.Fatal("unable to parse serviceID, make sure it's a numeric value")
			}
			expiration, err = strconv.ParseInt(ttl, 10, 64)
			if err != nil {
				expiration = 0
			}

			key, err := rootCmd.GrafanaSvc().CreateServiceAccountToken(serviceID, name, expiration)
			if err != nil {
				log.Fatal("unable to create api key", "err", err)
			} else {

				rootCmd.TableObj.AppendHeader(table.Row{"serviceID", "token_id", "name", "token"})
				rootCmd.TableObj.AppendRow(table.Row{serviceID, key.ID, key.Name, key.Key})
				rootCmd.TableObj.Render()
			}

			return nil
		},
	}
}
