package cmd

import (
	"github.com/netsage-project/grafana-dashboard-manager/config"
	"github.com/spf13/cobra"
)

var contextSet = &cobra.Command{
	Use:   "set",
	Short: "set context",
	Long:  `set contexts.`,
	Run: func(cmd *cobra.Command, args []string) {
		context, _ := cmd.Flags().GetString("context")
		config.SetContext(context)

	},
}

func init() {
	context.AddCommand(contextSet)
	contextSet.Flags().StringP("context", "c", "", "context")
	contextSet.MarkFlagRequired("context")
}
