package cmd

import (
	"fmt"

	"github.com/jedib0t/go-pretty/table"
	"github.com/netsage-project/grafana-dashboard-manager/apphelpers"
	"github.com/spf13/cobra"
)

var contextList = &cobra.Command{
	Use:   "list",
	Short: "List context",
	Long:  `List contexts.`,
	Run: func(cmd *cobra.Command, args []string) {
		tableObj.AppendHeader(table.Row{"context", "active"})
		contexts := apphelpers.GetContexts()
		activeContext := apphelpers.GetContext()
		for _, item := range contexts {
			active := false
			if item == activeContext {
				item = fmt.Sprintf("*%s", item)
				active = true
			}
			tableObj.AppendRow(table.Row{item, active})
		}

		tableObj.Render()
	},
}

func init() {
	context.AddCommand(contextList)
}
