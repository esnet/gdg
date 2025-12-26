package backup

import (
	"context"
	"log/slog"
	"os"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/jedib0t/go-pretty/v6/table"
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

// getConnectionTbWriter returns a table object for use with newConnectionsPermissionListCmd
func getFolderPermTblWriter() table.Writer {
	writer := table.NewWriter()
	writer.SetOutputMirror(os.Stdout)
	writer.SetStyle(table.StyleLight)
	writer.AppendHeader(table.Row{"folder ID", "folderUid", "folder Name", "nested path"}, table.RowConfig{AutoMerge: true})
	return writer
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

			slog.Info("Listing Folders for context", "context", rootCmd.ConfigSvc().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"folderUid", "folder ID", "folder Name", "UserID", "Team Name", "Role", "Permission Name"}, rowConfigAutoMerge)
			folders := rootCmd.GrafanaSvc().ListFolderPermissions(getFolderFilter(rootCmd.ConfigSvc()))

			if len(folders) == 0 {
				slog.Info("No folders found")
				return nil
			}
			for key, value := range folders {
				writer := getFolderPermTblWriter()
				writer.AppendRow(table.Row{key.UID, key.ID, key.Title, key.NestedPath})
				writer.Render()
				if len(value) > 0 {
					twConfigs := table.NewWriter()
					twConfigs.SetOutputMirror(os.Stdout)
					twConfigs.SetStyle(table.StyleDouble)
					twConfigs.AppendHeader(table.Row{"Folder UID", "UserID", "Team Name", "Role", "Permission Name"})
					for _, entry := range value {
						twConfigs.AppendRow(table.Row{key.UID, entry.UserLogin, entry.Team, entry.Role, entry.PermissionName})
					}
					twConfigs.Render()
				}
			}
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
			slog.Info("Downloading Folder Permissions for context", "context", rootCmd.ConfigSvc().GetContext())
			rootCmd.TableObj.AppendHeader(table.Row{"filename"})
			folders := rootCmd.GrafanaSvc().DownloadFolderPermissions(getFolderFilter(rootCmd.ConfigSvc()))
			slog.Info("Downloading folder permissions")

			if len(folders) == 0 {
				slog.Info("No folders found")
				return nil
			}
			for _, folder := range folders {
				rootCmd.TableObj.AppendRow(table.Row{folder})
			}
			rootCmd.Render(cd.CobraCommand, folders)
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
			slog.Info("Uploading folder permissions")
			rootCmd.TableObj.AppendHeader(table.Row{"file name"})
			folders := rootCmd.GrafanaSvc().UploadFolderPermissions(getFolderFilter(rootCmd.ConfigSvc()))

			if len(folders) == 0 {
				slog.Info("No folders found")
				return nil
			}
			for _, folder := range folders {
				rootCmd.TableObj.AppendRow(table.Row{folder})
			}
			rootCmd.Render(cd.CobraCommand, folders)
			return nil
		},
	}
}
