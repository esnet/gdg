package cmd

import (
	"github.com/esnet/gdg/api"
	"github.com/esnet/gdg/api/filters"
	"github.com/esnet/gdg/apphelpers"
	"github.com/jedib0t/go-pretty/table"
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
	return api.NewFolderFilter()

}

var foldersDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete Folders",
	Long:  `delete Folders`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Deleting all Folders for context: '%s'", apphelpers.GetContext())
		tableObj.AppendHeader(table.Row{"title"})

		folders := client.DeleteAllFolder(getFolderFilter())
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

var foldersExportCmd = &cobra.Command{
	Use:   "export",
	Short: "export Folders",
	Long:  `export Folders`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Listing Folders for context: '%s'", apphelpers.GetContext())
		tableObj.AppendHeader(table.Row{"file"})
		folders := client.ExportFolder(getFolderFilter())
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

var foldersImportCmd = &cobra.Command{
	Use:   "import",
	Short: "import Folders",
	Long:  `import Folders`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Listing Folders for context: '%s'", apphelpers.GetContext())
		tableObj.AppendHeader(table.Row{"file"})
		folders := client.ImportFolder(getFolderFilter())
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

var foldersListCmd = &cobra.Command{
	Use:   "list",
	Short: "list Folders",
	Long:  `list Folders`,
	Run: func(cmd *cobra.Command, args []string) {

		log.Infof("Listing Folders for context: '%s'", apphelpers.GetContext())
		tableObj.AppendHeader(table.Row{"id", "uid", "title"})
		folders := client.ListFolder(getFolderFilter())

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
	foldersCmd.AddCommand(foldersDeleteCmd)
	foldersCmd.AddCommand(foldersExportCmd)
	foldersCmd.AddCommand(foldersImportCmd)
	foldersCmd.AddCommand(foldersListCmd)
	foldersCmd.PersistentFlags().BoolVar(&useFolderFilters, "use-filters", false, "Default to false, but if passed then will only operate on the list of folders listed in the configuration file")

}
