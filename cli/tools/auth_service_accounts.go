package tools

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"sort"
	"strconv"
	"strings"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/spf13/cobra"
)

func newServiceAccountCmd() simplecobra.Commander {
	description := "Manage api service-account"
	return &support.SimpleCommand{
		NameP: "service-accounts",
		Short: description,
		Long:  description,
		CommandsList: []simplecobra.Commander{
			newServiceAccountTokensCmd(),
			newListServiceAccountCmd(),
			newClearServiceAccountsCmd(),
			newDeleteServiceAccountsCmd(),
			newServiceAccount(),
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
							formattedDate := token.Expiration.String()
							date, _ := token.Expiration.Value()
							if date.(string) == "0001-01-01T00:00:00.000Z" {
								formattedDate = "No Expiration"
							}
							rootCmd.TableObj.AppendRow(table.Row{"", "", "", "", token.ID, token.Name, formattedDate})
						}
					}
				}
				rootCmd.Render(cd.CobraCommand, apiKeys)
			}

			return nil
		},
	}
}

func newDeleteServiceAccountsCmd() simplecobra.Commander {
	description := "delete the given service account from grafana"
	return &support.SimpleCommand{
		NameP:        "delete",
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

			slog.Info("Deleting Service Accounts for context", "context", config.Config().GetGDGConfig().GetContext(),
				"serviceAccountId", id)
			err = rootCmd.GrafanaSvc().DeleteServiceAccount(id)
			rootCmd.TableObj.AppendHeader(table.Row{"type", "filename"})
			if err != nil {
				slog.Info("Unable to delete service account", "err", err.Error())
			} else {
				slog.Info("Service account has been removed", "serviceAccountId", id)
			}
			return nil
		},
	}
}

func newClearServiceAccountsCmd() simplecobra.Commander {
	description := "delete all Service Accounts from grafana"
	return &support.SimpleCommand{
		NameP:        "clear",
		Short:        description,
		Long:         description,
		CommandsList: []simplecobra.Commander{},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			savedFiles := rootCmd.GrafanaSvc().DeleteAllServiceAccounts()
			slog.Info("Delete Service Accounts for context", "context", config.Config().GetGDGConfig().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"type", "filename"})
			if len(savedFiles) == 0 {
				slog.Info("No Service Accounts found")
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

func newServiceAccount() simplecobra.Commander {
	description := "new <serviceName> <role> [ttl in seconds]"
	return &support.SimpleCommand{
		NameP:        "new",
		Short:        description,
		Long:         description,
		CommandsList: []simplecobra.Commander{},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("requires a key name and a role(%s) [ttl optional]", strings.Join(getBasicRoles(), ", "))
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

			if !validBasicRole(role) {
				log.Fatalf("Invalid role specified, '%s'.  Valid roles are:[%s]", role, strings.Join(getBasicRoles(), ", "))
			}
			serviceAcct, err := rootCmd.GrafanaSvc().CreateServiceAccount(name, role, expiration)
			if err != nil {
				log.Fatal("unable to create api key", "error", err)
			} else {

				rootCmd.TableObj.AppendHeader(table.Row{"id", "name", "role"})
				rootCmd.TableObj.AppendRow(table.Row{serviceAcct.ID, serviceAcct.Name, serviceAcct.Role})
				rootCmd.Render(cd.CobraCommand,
					map[string]any{"id": serviceAcct.ID, "name": serviceAcct.Name, "role": serviceAcct.Role})
			}
			return nil
		},
	}
}
