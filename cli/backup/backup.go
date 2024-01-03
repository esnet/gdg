package backup

import (
	"context"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/spf13/cobra"
)

func NewBackupCommand() simplecobra.Commander {
	description := "Manage entities that are backup up and updated via api"
	return &support.SimpleCommand{
		NameP: "backup",
		Short: description,
		Long: `Manage entities that are backup up and updated via api.  These utilities are mostly
limited to clear/delete, list, download and upload.  Any other functionality will be found under the tools.`,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"b"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
		InitCFunc: func(cd *simplecobra.Commandeer, r *support.RootCommand) error {
			support.InitConfiguration(cd.CobraCommand)
			r.GrafanaSvc().InitOrganizations()
			return nil
		},
		CommandsList: []simplecobra.Commander{
			newDashboardCommand(),
			newAlertNotificationsCommand(),
			newConnectionsCommand(),
			newFolderCommand(),
			newLibraryElementsCommand(),
			newOrganizationsCommand(),
			newTeamsCommand(),
			newUsersCommand(),
		},
	}

}
