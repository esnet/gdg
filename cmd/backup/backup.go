package backup

import (
	"github.com/esnet/gdg/cmd"
	"github.com/spf13/cobra"
)

// userCmd represents the version command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage entities that are backup up and updated via api",
	Long: `Manage entities that are backup up and updated via api.  These utilities are mostly 
limited to clear/delete, list, download and upload.  Any other functionality will be found under the tools.`,
	Aliases: []string{"b"},
}

func init() {
	cmd.RootCmd.AddCommand(backupCmd)
	cmd.GetGrafanaSvc().InitOrganizations()
}
