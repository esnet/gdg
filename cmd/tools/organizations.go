package tools

import (
	"errors"
	"github.com/esnet/gdg/cmd"
	"github.com/esnet/gdg/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
)

var orgCmd = &cobra.Command{
	Use:     "organizations",
	Aliases: []string{"org", "orgs"},
	Short:   "Manage Organizations",
	Long:    `Manage Grafana Organizations.`,
}

var setOrgCmd = &cobra.Command{
	Use:   "set",
	Short: "set <OrgId>, 0 removes filter",
	Long:  `set <OrgId>, 0 removes filter`,
	Args: func(command *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires an Org ID")
		}
		return nil
	},

	Run: func(command *cobra.Command, args []string) {
		ctx := args[0]
		orgId, err := strconv.ParseInt(ctx, 10, 64)
		if err != nil {
			log.Fatal("invalid Org ID, could not parse value to a numeric value")
		}
		err = cmd.GetGrafanaSvc().SetOrganization(orgId)
		if err != nil {
			log.WithError(err).Fatal("unable to set Org ID")
		}
		log.Infof("Succesfully set Org ID for context: %s", config.Config().AppConfig.GetContext())
	},
}

func init() {
	toolsCmd.AddCommand(orgCmd)
	orgCmd.AddCommand(setOrgCmd)
}
