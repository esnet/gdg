package tools

import (
	"github.com/esnet/gdg/cmd"
	"github.com/spf13/cobra"
)

// userCmd represents the version command
var toolsCmd = &cobra.Command{
	Use:     "tools",
	Short:   "A collection of tools to manage a grafana instance",
	Long:    `A collection of tools to manage a grafana instance`,
	Aliases: []string{"t"},
}

func init() {
	cmd.RootCmd.AddCommand(toolsCmd)
}
