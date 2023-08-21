package tools

import (
	"errors"
	cmd "github.com/esnet/gdg/cmd"
	"github.com/esnet/gdg/internal/config"
	"sort"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var serviceAcctsCmd = &cobra.Command{
	Use:     "service-accounts",
	Aliases: []string{"service-account", "svcAcct", "svcAccts", "svc"},
	Short:   "Manage api service-account",
	Long:    `Provides some utility to help the user manage their API token keys`,
}

var listServiceAcctsCmd = &cobra.Command{
	Use:   "list",
	Short: "list API Keys",
	Long:  `list API Keys`,
	Run: func(command *cobra.Command, args []string) {

		cmd.TableObj.AppendHeader(table.Row{"id", "service name", "role", "tokens", "token id", "token name", "expiration"})
		apiKeys := cmd.GetGrafanaSvc().ListServiceAccounts()
		sort.SliceStable(apiKeys, func(i, j int) bool {
			return apiKeys[i].ServiceAccount.ID < apiKeys[j].ServiceAccount.ID
		})
		if len(apiKeys) == 0 {
			log.Info("No apiKeys found")
		} else {
			for _, apiKey := range apiKeys {

				cmd.TableObj.AppendRow(table.Row{apiKey.ServiceAccount.ID, apiKey.ServiceAccount.Name, apiKey.ServiceAccount.Role, apiKey.ServiceAccount.Tokens})
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
						cmd.TableObj.AppendRow(table.Row{"", "", "", "", token.ID, token.Name, formattedDate})
					}
				}
			}
			cmd.TableObj.Render()
		}

	},
}

var deleteServiceAcctsTokensCmd = &cobra.Command{
	Use:   "clearTokens",
	Short: "delete all tokens for Service Account from grafana",
	Long:  `delete all tokens for Service Account from grafana`,
	Args: func(command *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a service account ID to be specified")
		}
		return nil
	},

	Run: func(command *cobra.Command, args []string) {
		idStr := args[0]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Fatalf("unable to parse %s as a valid numeric value", idStr)
		}

		log.Infof("Deleting Service Accounts Tokens for serviceID %d for context: '%s'", id, config.Config().AppConfig.GetContext())
		savedFiles := cmd.GetGrafanaSvc().DeleteServiceAccountTokens(id)
		cmd.TableObj.AppendHeader(table.Row{"serviceID", "type", "token_name"})
		if len(savedFiles) == 0 {
			log.Info("No Service Accounts tokens found")
		} else {
			for _, token := range savedFiles {
				cmd.TableObj.AppendRow(table.Row{id, "service token", token})
			}
			cmd.TableObj.Render()
		}
	},
}

var deleteServiceAcctsCmd = &cobra.Command{
	Use:   "clear",
	Short: "delete all Service Accounts from grafana",
	Long:  `delete all Service Accounts from grafana`,
	Run: func(command *cobra.Command, args []string) {

		savedFiles := cmd.GetGrafanaSvc().DeleteAllServiceAccounts()
		log.Infof("Delete Service Accounts for context: '%s'", config.Config().AppConfig.GetContext())
		cmd.TableObj.AppendHeader(table.Row{"type", "filename"})
		if len(savedFiles) == 0 {
			log.Info("No Service Accounts found")
		} else {
			for _, file := range savedFiles {
				cmd.TableObj.AppendRow(table.Row{"user", file})
			}
			cmd.TableObj.Render()
		}
	},
}

var newServiceAcctsCmd = &cobra.Command{
	Use:   "newService",
	Short: "newService <serviceName> <role> [ttl in seconds]",
	Long:  `newService <serviceName> <role> [ttl in seconds]`,
	Args: func(command *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("requires a key name and a role('admin','viewer','editor') [ttl optional] ")
		}
		return nil
	},
	Run: func(command *cobra.Command, args []string) {
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
		serviceAcct, err := cmd.GetGrafanaSvc().CreateServiceAccount(name, role, expiration)
		if err != nil {
			log.WithError(err).Fatal("unable to create api key")
		} else {

			cmd.TableObj.AppendHeader(table.Row{"id", "name", "role"})
			cmd.TableObj.AppendRow(table.Row{serviceAcct.ID, serviceAcct.Name, serviceAcct.Role})
			cmd.TableObj.Render()
		}

	},
}

var newServiceAcctsTokenCmd = &cobra.Command{
	Use:   "newToken",
	Short: "newToken <serviceAccountID> <name> [ttl in seconds]",
	Long:  `newToken <serviceAccountID> <name> [ttl in seconds]`,
	Args: func(command *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("requires a service-account ID and token name [ttl optional] ")
		}
		return nil
	},
	Run: func(command *cobra.Command, args []string) {
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

		key, err := cmd.GetGrafanaSvc().CreateServiceAccountToken(serviceID, name, expiration)
		if err != nil {
			log.WithError(err).Fatal("unable to create api key")
		} else {

			cmd.TableObj.AppendHeader(table.Row{"serviceID", "token_id", "name", "token"})
			cmd.TableObj.AppendRow(table.Row{serviceID, key.ID, key.Name, key.Key})
			cmd.TableObj.Render()
		}

	},
}

func init() {
	AuthCmd.AddCommand(serviceAcctsCmd)
	serviceAcctsCmd.AddCommand(listServiceAcctsCmd)
	serviceAcctsCmd.AddCommand(deleteServiceAcctsCmd)
	serviceAcctsCmd.AddCommand(deleteServiceAcctsTokensCmd)
	serviceAcctsCmd.AddCommand(newServiceAcctsTokenCmd)
	serviceAcctsCmd.AddCommand(newServiceAcctsCmd)
}
