package backup

import (
	"context"

	"github.com/bep/simplecobra"
	domain2 "github.com/esnet/gdg/cli/domain"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/spf13/cobra"
)

func NewBackupCommand() simplecobra.Commander {
	description := "Manage entities that are backup up and updated via api"
	return &domain2.SimpleCommand{
		NameP: "backup",
		Short: description,
		Long: `Manage entities that are backup up and updated via api.  These utilities are mostly
limited to clear/delete, list, download and upload.  Any other functionality will be found under the tools.`,
		WithCFunc: func(cmd *cobra.Command, r *domain2.RootCommand) {
			cmd.Aliases = []string{"b"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain2.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
		InitCFunc: func(cd *simplecobra.Commandeer, r *domain2.RootCommand) error {
			configOverride, _ := cd.CobraCommand.Flags().GetString("config")
			contextOverride, _ := cd.CobraCommand.Flags().GetString("context")
			r.LoadConfig(configOverride, contextOverride)
			r.GrafanaSvc().Login()
			r.GrafanaSvc().InitOrganizations()
			return nil
		},
		CommandsList: []simplecobra.Commander{
			newDashboardCommand(),
			newConnectionsCommand(),
			newFolderCommand(),
			newLibraryElementsCommand(),
			newOrganizationsCommand(),
			newTeamsCommand(),
			newUsersCommand(),
			newAlertingCommand(),
		},
	}
}

// GetOrganizationName wrapper for verbose version below.
func GetOrganizationName(cfg *config_domain.GDGAppConfiguration) string {
	return cfg.GetDefaultGrafanaConfig().GetOrganizationName()
}
