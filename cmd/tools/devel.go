package tools

import (
	"github.com/esnet/gdg/cmd"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var develCmd = &cobra.Command{
	Use:   "devel",
	Short: "Developer Tooling",
	Long:  `Developer Tooling`,
}

var CompletionCmd = &cobra.Command{
	Use:                   "completion [bash|zsh|fish|powershell]",
	Short:                 "Generate completion script",
	Long:                  "To load completions",
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(command *cobra.Command, args []string) {
		var err error
		switch args[0] {
		case "bash":
			err = cmd.RootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			err = cmd.RootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			err = cmd.RootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			err = cmd.RootCmd.GenPowerShellCompletion(os.Stdout)
		}
		if err != nil {
			log.Error("Failed to generation shell completion.")
		}
	},
}

var serverInfo = &cobra.Command{
	Use:   "srvinfo",
	Short: "server health info",
	Long:  `server health info`,
	Run: func(command *cobra.Command, args []string) {
		result := cmd.GetGrafanaSvc().GetServerInfo()
		for key, value := range result {
			log.Infof("%s:  %s", key, value)
		}
	},
}

func init() {
	toolsCmd.AddCommand(develCmd)
	develCmd.AddCommand(CompletionCmd)
	develCmd.AddCommand(serverInfo)
}
