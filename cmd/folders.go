package cmd

import (
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
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Deleting all Folders for context: '%s'", config.Config().AppConfig.GetContext())
		tableObj.AppendHeader(table.Row{"title"})

		folders := grafanaSvc.DeleteAllFolder(getFolderFilter())
		if len(folders) == 0 {
			log.Info("No Folders found")
		} else {
			for _, folder := range folders {
				tableObj.AppendRow(table.Row{folder})
			}
			tableObj.Render()
		}

	},
}

var uploadFoldersCmd = &cobra.Command{
	Use:     "upload",
	Short:   "upload Folders to grafana",
	Long:    `upload Folders to grafana`,
	Aliases: []string{"export", "u"},
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Listing Folders for context: '%s'", config.Config().AppConfig.GetContext())
		tableObj.AppendHeader(table.Row{"file"})
		folders := grafanaSvc.ExportFolder(getFolderFilter())
		if len(folders) == 0 {
			log.Info("No folders found")
		} else {
			for _, folder := range folders {
				tableObj.AppendRow(table.Row{folder})
			}
			tableObj.Render()
		}

	},
}

var downloadFoldersCmd = &cobra.Command{
	Use:     "download",
	Short:   "download Folders",
	Long:    `download Folders`,
	Aliases: []string{"import", "d"},
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Listing Folders for context: '%s'", config.Config().AppConfig.GetContext())
		tableObj.AppendHeader(table.Row{"file"})
		folders := grafanaSvc.ImportFolder(getFolderFilter())
		if len(folders) == 0 {
			log.Info("No folders found")
		} else {
			for _, folder := range folders {
				tableObj.AppendRow(table.Row{folder})
			}
			tableObj.Render()
		}

	},
}

var listFoldersCmd = &cobra.Command{
	Use:     "list",
	Short:   "list Folders",
	Long:    `list Folders`,
	Aliases: []string{"l"},
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Listing Folders for context: '%s'", config.Config().AppConfig.GetContext())
		tableObj.AppendHeader(table.Row{"id", "uid", "title"})
		folders := grafanaSvc.ListFolder(getFolderFilter())

		if len(folders) == 0 {
			log.Info("No folders found")
		} else {
			for _, folder := range folders {
				tableObj.AppendRow(table.Row{folder.ID, folder.UID, folder.Title})
			}
			tableObj.Render()
		}

	},
}

func init() {
	rootCmd.AddCommand(foldersCmd)
	foldersCmd.AddCommand(deleteFoldersCmd)
	foldersCmd.AddCommand(uploadFoldersCmd)
	foldersCmd.AddCommand(downloadFoldersCmd)
	foldersCmd.AddCommand(listFoldersCmd)
	foldersCmd.AddCommand(listFoldersPermissionsCmd)
	foldersCmd.PersistentFlags().BoolVar(&useFolderFilters, "use-filters", false, "Default to false, but if passed then will only operate on the list of folders listed in the configuration file")

}
