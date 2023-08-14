package cmd

import (
	"github.com/esnet/gdg/internal/config"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var folderPermissionCmd = &cobra.Command{
	Use:     "permission",
	Aliases: []string{"p", "permissions"},
	Short:   "Folders Permission",
	Long:    `Folders Permission`,
}

var listFoldersPermissionsCmd = &cobra.Command{
	Use:   "list",
	Short: "list Folders Permissions",
	Long:  `list Folders Permissions`,
	Run: func(cmd *cobra.Command, args []string) {
		rowConfigAutoMerge := table.RowConfig{AutoMerge: true}

		log.Infof("Listing Folders for context: '%s'", config.Config().GetAppConfig().GetContext())
		tableObj.AppendHeader(table.Row{"folder ID", "folderUid", "folder Name", "UserID", "Team Name", "Role", "Permission Name"}, rowConfigAutoMerge)
		folders := grafanaSvc.ListFolderPermissions(getFolderFilter())

		if len(folders) == 0 {
			log.Info("No folders found")
			return
		}
		for key, value := range folders {
			tableObj.AppendRow(table.Row{key.ID, key.UID, key.Title})
			for _, entry := range value {
				tableObj.AppendRow(table.Row{"", "", "    PERMISSION--->", entry.UserLogin, entry.Team, entry.Role, entry.PermissionName}, rowConfigAutoMerge)
			}
		}
		tableObj.Render()

	},
}

var downloadFoldersPermissionsCmd = &cobra.Command{
	Use:     "download",
	Short:   "download Folders Permissions",
	Long:    `downloadFolders Permissions`,
	Aliases: []string{"import"},
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("import Folders for context: '%s'", config.Config().GetAppConfig().GetContext())
		tableObj.AppendHeader(table.Row{"filename"})
		folders := grafanaSvc.ImportFolderPermissions(getFolderFilter())
		//_ = folders
		log.Infof("Downloading folder permissions")

		if len(folders) == 0 {
			log.Info("No folders found")
			return
		}
		for _, folder := range folders {
			tableObj.AppendRow(table.Row{folder})
		}
		tableObj.Render()

	},
}

var uploadFoldersPermissionsCmd = &cobra.Command{
	Use:     "upload",
	Short:   "upload Folders Permissions",
	Long:    `uploadFolders Permissions`,
	Aliases: []string{"export"},
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Uploading folder permissions")
		tableObj.AppendHeader(table.Row{"file name"})
		folders := grafanaSvc.ExportFolderPermissions(getFolderFilter())

		if len(folders) == 0 {
			log.Info("No folders found")
			return
		}
		for _, folder := range folders {
			tableObj.AppendRow(table.Row{folder})
		}
		tableObj.Render()

	},
}

func init() {
	foldersCmd.AddCommand(folderPermissionCmd)
	folderPermissionCmd.AddCommand(listFoldersPermissionsCmd)
	folderPermissionCmd.AddCommand(downloadFoldersPermissionsCmd)
	folderPermissionCmd.AddCommand(uploadFoldersPermissionsCmd)

}
