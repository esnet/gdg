package cmd

import (
	"fmt"
	"strings"

	"github.com/esnet/gdg/apphelpers"
	"github.com/jedib0t/go-pretty/table"
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
			if item == strings.ToLower(activeContext) {
				item = fmt.Sprintf("*%s", activeContext)
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
