package backup

import (
	"context"
	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cmd/support"
	"github.com/esnet/gdg/internal/config"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newFolderPermissionCommand() simplecobra.Commander {
	description := "Folder Permissions"
	return &support.SimpleCommand{
		NameP: "permission",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"p", "permissions"}
		},
		CommandsList: []simplecobra.Commander{
			newFolderPermissionListCmd(),
			newFolderPermissionUploadCmd(),
			newFolderPermissionDownloadCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newFolderPermissionListCmd() simplecobra.Commander {
	description := "list Folder Permissions"
	return &support.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"l"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rowConfigAutoMerge := table.RowConfig{AutoMerge: true}

			log.Infof("Listing Folders for context: '%s'", config.Config().GetAppConfig().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"folder ID", "folderUid", "folder Name", "UserID", "Team Name", "Role", "Permission Name"}, rowConfigAutoMerge)
			folders := rootCmd.GrafanaSvc().ListFolderPermissions(getFolderFilter())

			if len(folders) == 0 {
				log.Info("No folders found")
				return nil
			}
			for key, value := range folders {
				rootCmd.TableObj.AppendRow(table.Row{key.ID, key.UID, key.Title})
				for _, entry := range value {
					rootCmd.TableObj.AppendRow(table.Row{"", "", "    PERMISSION--->", entry.UserLogin, entry.Team, entry.Role, entry.PermissionName}, rowConfigAutoMerge)
				}
			}
			rootCmd.TableObj.Render()
			return nil
		},
	}
}
func newFolderPermissionDownloadCmd() simplecobra.Commander {
	description := "download Folders Permissions"
	return &support.SimpleCommand{
		NameP: "download",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			log.Infof("Downloading Folder Permissions for context: '%s'", config.Config().GetAppConfig().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"filename"})
			folders := rootCmd.GrafanaSvc().DownloadFolderPermissions(getFolderFilter())
			log.Infof("Downloading folder permissions")

			if len(folders) == 0 {
				log.Info("No folders found")
				return nil
			}
			for _, folder := range folders {
				rootCmd.TableObj.AppendRow(table.Row{folder})
			}
			rootCmd.TableObj.Render()
			return nil
		},
	}
}
func newFolderPermissionUploadCmd() simplecobra.Commander {
	description := "upload Folders Permissions"
	return &support.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			log.Infof("Uploading folder permissions")
			rootCmd.TableObj.AppendHeader(table.Row{"file name"})
			folders := rootCmd.GrafanaSvc().UploadFolderPermissions(getFolderFilter())

			if len(folders) == 0 {
				log.Info("No folders found")
				return nil
			}
			for _, folder := range folders {
				rootCmd.TableObj.AppendRow(table.Row{folder})
			}
			rootCmd.TableObj.Render()
			return nil
		},
	}
}
