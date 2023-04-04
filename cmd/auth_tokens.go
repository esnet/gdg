package cmd

import (
	"errors"
	"github.com/esnet/gdg/internal/apphelpers"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
	"sort"
	"strconv"
)

var tokensCmd = &cobra.Command{
	Use:     "tokens",
	Aliases: []string{"token", "apikeys"},
	Short:   "Manage api tokens",
	Long:    `Provides some utility to help the user manage their API token keys`,
}

var listTokensCmd = &cobra.Command{
	Use:   "list",
	Short: "list API Keys",
	Long:  `list API Keys`,
	Run: func(cmd *cobra.Command, args []string) {

		tableObj.AppendHeader(table.Row{"id", "name", "role", "expiration"})
		apiKeys := grafanaSvc.ListAPIKeys()
		sort.SliceStable(apiKeys, func(i, j int) bool {
			return apiKeys[i].ID < apiKeys[j].ID
		})
		if len(apiKeys) == 0 {
			log.Info("No apiKeys found")
		} else {
			for _, apiKey := range apiKeys {
				var formattedDate string = apiKey.Expiration.String()
				date, _ := apiKey.Expiration.Value()
				if date.(string) == "0001-01-01T00:00:00.000Z" {
					formattedDate = "No Expiration"
				}

				tableObj.AppendRow(table.Row{apiKey.ID, apiKey.Name, apiKey.Role, formattedDate})
			}
			tableObj.Render()
		}

	},
}

var deleteTokensCmd = &cobra.Command{
	Use:   "clear",
	Short: "delete all Tokens",
	Long:  `delete all Tokens`,
	Run: func(cmd *cobra.Command, args []string) {

		savedFiles := grafanaSvc.DeleteAllTokens()
		log.Infof("Delete Tokens for context: '%s'", apphelpers.GetContext())
		tableObj.AppendHeader(table.Row{"type", "filename"})
		if len(savedFiles) == 0 {
			log.Info("No Tokens found")
		} else {
			for _, file := range savedFiles {
				tableObj.AppendRow(table.Row{"user", file})
			}
			tableObj.Render()
		}
	},
}

var newTokensCmd = &cobra.Command{
	Use:   "new",
	Short: "new <name> <role> [ttl in seconds]",
	Long:  `new <name> <role> [ttl in seconds ]`,
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
		key, err := grafanaSvc.CreateAPIKey(name, role, expiration)
		if err != nil {
			log.WithError(err).Fatal("unable to create api key")
		} else {

			tableObj.AppendHeader(table.Row{"id", "name", "token"})
			tableObj.AppendRow(table.Row{key.ID, key.Name, key.Key})
			tableObj.Render()
		}

	},
}

func init() {
	tokensCmd.AddCommand(listTokensCmd)
	tokensCmd.AddCommand(deleteTokensCmd)
	tokensCmd.AddCommand(newTokensCmd)
}
