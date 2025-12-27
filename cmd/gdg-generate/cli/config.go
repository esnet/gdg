package cli

import (
	"fmt"
	"log"
	"log/slog"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var showConfigCmd = &cobra.Command{
	Use:     "config",
	Short:   "Show current templates configuration",
	Long:    `Show current templates configuration`,
	Aliases: []string{"cfg"},
	Run: func(cmd *cobra.Command, args []string) {
		data, err := yaml.Marshal(templateConfig)
		if err != nil {
			log.Fatalf("unable to load template configuration: %v", err)
		}
		slog.Info("Configuration",
			slog.String("template-config", tplCfgFile),
			slog.String("gdg-config", cfgFile))
		fmt.Println(string(data))
	},
}

func init() {
	rootCmd.AddCommand(showConfigCmd)
}
