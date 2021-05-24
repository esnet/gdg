package cmd

import (
	"github.com/jedib0t/go-pretty/table"
	"github.com/netsage-project/grafana-dashboard-manager/config"
	"github.com/spf13/cobra"
)

var contextList = &cobra.Command{
	Use:   "list",
	Short: "List context",
	Long:  `List contexts.`,
	Run: func(cmd *cobra.Command, args []string) {
		tableObj.AppendHeader(table.Row{"context", "active"})
		contexts := config.GetContexts()
		activeContext := config.GetContext()
		for _, item := range contexts {
			tableObj.AppendRow(table.Row{item, item == activeContext})
		}

		tableObj.Render()
	},
}

func init() {
	context.AddCommand(contextList)
}
