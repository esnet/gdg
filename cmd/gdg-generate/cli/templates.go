package cli

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var tplCmd = &cobra.Command{
	Use:     "template",
	Aliases: []string{"tpl", "templates"},
	Short:   "Templating Utilities",
	Long:    `Templating Utilities`,
}

var listTemplatesCmd = &cobra.Command{
	Use:   "list",
	Short: "List current templates",
	Long:  `List current templates`,
	Run: func(cmd *cobra.Command, args []string) {
		templates := template.ListTemplates()
		slog.Info("Available templates for current configuration",
			slog.String("template-config", tplCfgFile),
			slog.String("gdg-config", cfgFile))
		for ndx, t := range templates {
			slog.Info(fmt.Sprintf("%d: %s", ndx+1, t))
		}
	},
}

var generateTemplatesCmd = &cobra.Command{
	Use:     "generate",
	Aliases: []string{},
	Short:   "Generate current templates",
	Long:    `Generate current templates`,
	Run: func(cmd *cobra.Command, args []string) {
		templateFilter, _ := cmd.Flags().GetString("template")
		payload, err := template.Generate(templateFilter)
		if err != nil {
			log.Fatal("Failed to generate templates", slog.Any("err", err))
		}

		tableObj := table.NewWriter()
		tableObj.SetOutputMirror(os.Stdout)
		tableObj.SetStyle(table.StyleLight)

		tableObj.AppendHeader(table.Row{"Template Name", "Output"})
		count := 0
		for key, val := range payload {
			count += len(val)
			for _, file := range val {
				tableObj.AppendRow(table.Row{key, file})
			}
		}
		slog.Info("Generate dashboards for the given templates",
			slog.Any("template-count", len(payload)),
			slog.Any("dashboard-count", count))
		tableObj.Render()
	},
}

func init() {
	rootCmd.AddCommand(tplCmd)
	tplCmd.AddCommand(listTemplatesCmd)
	tplCmd.AddCommand(generateTemplatesCmd)
	generateTemplatesCmd.PersistentFlags().StringP("template", "t", "", "Specify template name, optional.  Default is to operate on all configured templates that are found.")
}
