package tools

import (
	"errors"
	cmd "github.com/esnet/gdg/cmd"
	"github.com/esnet/gdg/internal/config"
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
	Run: func(command *cobra.Command, args []string) {

		cmd.TableObj.AppendHeader(table.Row{"id", "name", "role", "expiration"})
		apiKeys := cmd.GetGrafanaSvc().ListAPIKeys()
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

				cmd.TableObj.AppendRow(table.Row{apiKey.ID, apiKey.Name, apiKey.Role, formattedDate})
			}
			cmd.TableObj.Render()
		}

	},
}

var deleteTokensCmd = &cobra.Command{
	Use:   "clear",
	Short: "delete all Tokens from grafana",
	Long:  `delete all Tokens from grafana`,
	Run: func(command *cobra.Command, args []string) {

		savedFiles := cmd.GetGrafanaSvc().DeleteAllTokens()
		log.Infof("Delete Tokens for context: '%s'", config.Config().AppConfig.GetContext())
		cmd.TableObj.AppendHeader(table.Row{"type", "filename"})
		if len(savedFiles) == 0 {
			log.Info("No Tokens found")
		} else {
			for _, file := range savedFiles {
				cmd.TableObj.AppendRow(table.Row{"user", file})
			}
			cmd.TableObj.Render()
		}
	},
}

var newTokensCmd = &cobra.Command{
	Use:   "new",
	Short: "new <name> <role> [ttl in seconds]",
	Long:  `new <name> <role> [ttl in seconds ]`,
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
		key, err := cmd.GetGrafanaSvc().CreateAPIKey(name, role, expiration)
		if err != nil {
			log.WithError(err).Fatal("unable to create api key")
		} else {

			cmd.TableObj.AppendHeader(table.Row{"id", "name", "token"})
			cmd.TableObj.AppendRow(table.Row{key.ID, key.Name, key.Key})
			cmd.TableObj.Render()
		}

	},
}

func init() {
	tokensCmd.AddCommand(listTokensCmd)
	tokensCmd.AddCommand(deleteTokensCmd)
	tokensCmd.AddCommand(newTokensCmd)
}
