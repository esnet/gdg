package backup

import (
	"github.com/esnet/gdg/cmd"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var useFolderFilters bool

var foldersCmd = &cobra.Command{
	Use:     "folders",
	Aliases: []string{"fld", "folder"},
	Short:   "Folders Tooling",
	Long:    `Folders Tooling`,
}

func getFolderFilter() filters.Filter {
	if !useFolderFilters {
		return nil
	}
	return service.NewFolderFilter()

}

var deleteFoldersCmd = &cobra.Command{
	Use:     "clear",
	Aliases: []string{"delete"},
	Short:   "delete Folders from grafana",
	Long:    `delete Folders from grafana`,
	Run: func(command *cobra.Command, args []string) {

		log.Infof("Deleting all Folders for context: '%s'", config.Config().AppConfig.GetContext())
		cmd.TableObj.AppendHeader(table.Row{"title"})

		folders := cmd.GetGrafanaSvc().DeleteAllFolders(getFolderFilter())
		if len(folders) == 0 {
			log.Info("No Folders found")
		} else {
			for _, folder := range folders {
				cmd.TableObj.AppendRow(table.Row{folder})
			}
			cmd.TableObj.Render()
		}

	},
}

var uploadFoldersCmd = &cobra.Command{
	Use:     "upload",
	Short:   "upload Folders to grafana",
	Long:    `upload Folders to grafana`,
	Aliases: []string{"u"},
	Run: func(command *cobra.Command, args []string) {

		log.Infof("Listing Folders for context: '%s'", config.Config().AppConfig.GetContext())
		cmd.TableObj.AppendHeader(table.Row{"file"})
		folders := cmd.GetGrafanaSvc().UploadFolders(getFolderFilter())
		if len(folders) == 0 {
			log.Info("No folders found")
		} else {
			for _, folder := range folders {
				cmd.TableObj.AppendRow(table.Row{folder})
			}
			cmd.TableObj.Render()
		}

	},
}

var downloadFoldersCmd = &cobra.Command{
	Use:     "download",
	Short:   "download Folders",
	Long:    `download Folders`,
	Aliases: []string{"d"},
	Run: func(command *cobra.Command, args []string) {

		log.Infof("Listing Folders for context: '%s'", config.Config().AppConfig.GetContext())
		cmd.TableObj.AppendHeader(table.Row{"file"})
		folders := cmd.GetGrafanaSvc().DownloadFolders(getFolderFilter())
		if len(folders) == 0 {
			log.Info("No folders found")
		} else {
			for _, folder := range folders {
				cmd.TableObj.AppendRow(table.Row{folder})
			}
			cmd.TableObj.Render()
		}

	},
}

var listFoldersCmd = &cobra.Command{
	Use:     "list",
	Short:   "list Folders",
	Long:    `list Folders`,
	Aliases: []string{"l"},
	Run: func(command *cobra.Command, args []string) {

		log.Infof("Listing Folders for context: '%s'", config.Config().AppConfig.GetContext())
		cmd.TableObj.AppendHeader(table.Row{"id", "uid", "title"})
		folders := cmd.GetGrafanaSvc().ListFolder(getFolderFilter())

		if len(folders) == 0 {
			log.Info("No folders found")
		} else {
			for _, folder := range folders {
				cmd.TableObj.AppendRow(table.Row{folder.ID, folder.UID, folder.Title})
			}
			cmd.TableObj.Render()
		}

	},
}

func init() {
	backupCmd.AddCommand(foldersCmd)
	foldersCmd.AddCommand(listFoldersCmd)
	foldersCmd.AddCommand(downloadFoldersCmd)
	foldersCmd.AddCommand(deleteFoldersCmd)
	foldersCmd.AddCommand(uploadFoldersCmd)
	foldersCmd.AddCommand(listFoldersPermissionsCmd)
	foldersCmd.PersistentFlags().BoolVar(&useFolderFilters, "use-filters", false, "Default to false, but if passed then will only operate on the list of folders listed in the configuration file")

}
