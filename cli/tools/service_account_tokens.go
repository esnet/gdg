package tools

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"strconv"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func newServiceAccountTokensCmd() simplecobra.Commander {
	description := "Manage api service-account tokens"
	return &support.SimpleCommand{
		NameP: "tokens",
		Short: description,
		Long:  description,
		CommandsList: []simplecobra.Commander{
			newDeleteServiceAccountTokensCmd(),
			newServiceAccountTokenCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"token"}
		},
	}
}

func newDeleteServiceAccountTokensCmd() simplecobra.Commander {
	description := "clear <serviceAccountID>, removes all tokens from service account"
	return &support.SimpleCommand{
		NameP:        "clear",
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
				"context", rootCmd.ConfigSvc().GetContext())
			savedFiles := rootCmd.GrafanaSvc().DeleteServiceAccountTokens(id)
			rootCmd.TableObj.AppendHeader(table.Row{"serviceID", "type", "token_name"})
			if len(savedFiles) == 0 {
				slog.Info("No Service Accounts tokens found")
			} else {
				for _, token := range savedFiles {
					rootCmd.TableObj.AppendRow(table.Row{id, "service token", token})
				}
				rootCmd.Render(cd.CobraCommand, savedFiles)
			}
			return nil
		},
	}
}

func newServiceAccountTokenCmd() simplecobra.Commander {
	description := "new <serviceAccountID> <name> [ttl in seconds]"
	return &support.SimpleCommand{
		NameP:        "new",
		Short:        description,
		Long:         description,
		CommandsList: []simplecobra.Commander{},
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
				rootCmd.Render(cd.CobraCommand,
					map[string]any{
						"serviceID": serviceID,
						"token_id":  key.ID,
						"name":      key.Name,
						"token":     key.Key,
					})
			}

			return nil
		},
	}
}
