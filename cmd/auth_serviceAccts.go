package cmd

import (
	"errors"
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
	Run: func(cmd *cobra.Command, args []string) {

		tableObj.AppendHeader(table.Row{"id", "service name", "role", "tokens", "token id", "token name", "expiration"})
		apiKeys := grafanaSvc.ListServiceAccounts()
		sort.SliceStable(apiKeys, func(i, j int) bool {
			return apiKeys[i].ServiceAccount.ID < apiKeys[j].ServiceAccount.ID
		})
		if len(apiKeys) == 0 {
			log.Info("No apiKeys found")
		} else {
			for _, apiKey := range apiKeys {

				tableObj.AppendRow(table.Row{apiKey.ServiceAccount.ID, apiKey.ServiceAccount.Name, apiKey.ServiceAccount.Role, apiKey.ServiceAccount.Tokens})
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
						tableObj.AppendRow(table.Row{"", "", "", "", token.ID, token.Name, formattedDate})
					}
				}
			}
			tableObj.Render()
		}

	},
}

var deleteServiceAcctsTokensCmd = &cobra.Command{
	Use:   "clearTokens",
	Short: "delete all tokens for Service Account from grafana",
	Long:  `delete all tokens for Service Account from grafana`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a service account ID to be specified")
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		idStr := args[0]
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Fatalf("unable to parse %s as a valid numeric value", idStr)
		}

		log.Infof("Deleting Service Accounts Tokens for serviceID %d for context: '%s'", id, config.Config().AppConfig.GetContext())
		savedFiles := grafanaSvc.DeleteServiceAccountTokens(id)
		tableObj.AppendHeader(table.Row{"serviceID", "type", "token_name"})
		if len(savedFiles) == 0 {
			log.Info("No Service Accounts tokens found")
		} else {
			for _, token := range savedFiles {
				tableObj.AppendRow(table.Row{id, "service token", token})
			}
			tableObj.Render()
		}
	},
}

var deleteServiceAcctsCmd = &cobra.Command{
	Use:   "clear",
	Short: "delete all Service Accounts from grafana",
	Long:  `delete all Service Accounts from grafana`,
	Run: func(cmd *cobra.Command, args []string) {

		savedFiles := grafanaSvc.DeleteAllServiceAccounts()
		log.Infof("Delete Service Accounts for context: '%s'", config.Config().AppConfig.GetContext())
		tableObj.AppendHeader(table.Row{"type", "filename"})
		if len(savedFiles) == 0 {
			log.Info("No Service Accounts found")
		} else {
			for _, file := range savedFiles {
				tableObj.AppendRow(table.Row{"user", file})
			}
			tableObj.Render()
		}
	},
}

var newServiceAcctsCmd = &cobra.Command{
	Use:   "newService",
	Short: "newService <serviceName> <role> [ttl in seconds]",
	Long:  `newService <serviceName> <role> [ttl in seconds]`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("requires a key name and a role('admin','viewer','editor') [ttl optional] ")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
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
		serviceAcct, err := grafanaSvc.CreateServiceAccount(name, role, expiration)
		if err != nil {
			log.WithError(err).Fatal("unable to create api key")
		} else {

			tableObj.AppendHeader(table.Row{"id", "name", "role"})
			tableObj.AppendRow(table.Row{serviceAcct.ID, serviceAcct.Name, serviceAcct.Role})
			tableObj.Render()
		}

	},
}

var newServiceAcctsTokenCmd = &cobra.Command{
	Use:   "newToken",
	Short: "newToken <serviceAccountID> <name> [ttl in seconds]",
	Long:  `newToken <serviceAccountID> <name> [ttl in seconds]`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("requires a service-account ID and token name [ttl optional] ")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
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

		key, err := grafanaSvc.CreateServiceAccountToken(serviceID, name, expiration)
		if err != nil {
			log.WithError(err).Fatal("unable to create api key")
		} else {

			tableObj.AppendHeader(table.Row{"serviceID", "token_id", "name", "token"})
			tableObj.AppendRow(table.Row{serviceID, key.ID, key.Name, key.Key})
			tableObj.Render()
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
