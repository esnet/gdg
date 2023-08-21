package backup

import (
	"github.com/esnet/gdg/cmd"
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
	Run: func(command *cobra.Command, args []string) {
		rowConfigAutoMerge := table.RowConfig{AutoMerge: true}

		log.Infof("Listing Folders for context: '%s'", config.Config().GetAppConfig().GetContext())
		cmd.TableObj.AppendHeader(table.Row{"folder ID", "folderUid", "folder Name", "UserID", "Team Name", "Role", "Permission Name"}, rowConfigAutoMerge)
		folders := cmd.GetGrafanaSvc().ListFolderPermissions(getFolderFilter())

		if len(folders) == 0 {
			log.Info("No folders found")
			return
		}
		for key, value := range folders {
			cmd.TableObj.AppendRow(table.Row{key.ID, key.UID, key.Title})
			for _, entry := range value {
				cmd.TableObj.AppendRow(table.Row{"", "", "    PERMISSION--->", entry.UserLogin, entry.Team, entry.Role, entry.PermissionName}, rowConfigAutoMerge)
			}
		}
		cmd.TableObj.Render()

	},
}

var downloadFoldersPermissionsCmd = &cobra.Command{
	Use:   "download",
	Short: "download Folders Permissions",
	Long:  `downloadFolders Permissions`,
	Run: func(command *cobra.Command, args []string) {
		log.Infof("import Folders for context: '%s'", config.Config().GetAppConfig().GetContext())
		cmd.TableObj.AppendHeader(table.Row{"filename"})
		folders := cmd.GetGrafanaSvc().DownloadFolderPermissions(getFolderFilter())
		//_ = folders
		log.Infof("Downloading folder permissions")

		if len(folders) == 0 {
			log.Info("No folders found")
			return
		}
		for _, folder := range folders {
			cmd.TableObj.AppendRow(table.Row{folder})
		}
		cmd.TableObj.Render()

	},
}

var uploadFoldersPermissionsCmd = &cobra.Command{
	Use:   "upload",
	Short: "upload Folders Permissions",
	Long:  `uploadFolders Permissions`,
	Run: func(command *cobra.Command, args []string) {
		log.Infof("Uploading folder permissions")
		cmd.TableObj.AppendHeader(table.Row{"file name"})
		folders := cmd.GetGrafanaSvc().UploadFolderPermissions(getFolderFilter())

		if len(folders) == 0 {
			log.Info("No folders found")
			return
		}
		for _, folder := range folders {
			cmd.TableObj.AppendRow(table.Row{folder})
		}
		cmd.TableObj.Render()

	},
}

func init() {
	foldersCmd.AddCommand(folderPermissionCmd)
	folderPermissionCmd.AddCommand(listFoldersPermissionsCmd)
	folderPermissionCmd.AddCommand(downloadFoldersPermissionsCmd)
	folderPermissionCmd.AddCommand(uploadFoldersPermissionsCmd)

}
