package cmd

import (
	"errors"
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/jedib0t/go-pretty/v6/table"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

var context = &cobra.Command{
	Use:     "contexts",
	Aliases: []string{"ctx", "context"},
	Short:   "Manage Context configuration",
	Long:    `Manage Context configuration which allows multiple grafana configs to be used.`,
}

var contextClear = &cobra.Command{
	Use:   "clear",
	Short: "clear all context",
	Long:  `clear all contexts`,
	Run: func(cmd *cobra.Command, args []string) {
		config.Config().ClearContexts()
		log.Info("Successfully deleted all configured contexts")
	},
}

var contextCopy = &cobra.Command{
	Use:     "copy",
	Short:   "copy context <src> <dest>",
	Long:    `copy contexts  <src> <dest>`,
	Aliases: []string{"cp"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("requires a src and destination argument")
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		src := args[0]
		dest := args[1]
		config.Config().CopyContext(src, dest)
	},
}

var contextDelete = &cobra.Command{
	Use:     "delete",
	Short:   "delete context <context>",
	Long:    `delete context <context>.`,
	Aliases: []string{"del"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a context argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := args[0]
		config.Config().DeleteContext(ctx)
		log.Infof("Successfully deleted context %s", ctx)
	},
}

var contextList = &cobra.Command{
	Use:   "list",
	Short: "List context",
	Long:  `List contexts.`,
	Run: func(cmd *cobra.Command, args []string) {
		tableObj.AppendHeader(table.Row{"context", "active"})
		contexts := config.Config().GetAppConfig().GetContexts()
		activeContext := config.Config().GetAppConfig().GetContext()
		for key, _ := range contexts {
			active := false
			if key == strings.ToLower(activeContext) {
				key = fmt.Sprintf("*%s", activeContext)
				active = true
			}
			tableObj.AppendRow(table.Row{key, active})
		}

		tableObj.Render()
	},
}

var contextNew = &cobra.Command{
	Use:   "new",
	Short: "new <context>",
	Long:  `new <context>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a context name")
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		ctx := args[0]
		config.Config().NewContext(ctx)
	},
}

var contextSet = &cobra.Command{
	Use:   "set",
	Short: "set <context>",
	Long:  `set <contexts>`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a context argument")
		}
		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		context := args[0]
		config.Config().ChangeContext(context)

	},
}

var contextShow = &cobra.Command{
	Use:   "show",
	Short: "show optional[context]",
	Long:  `show contexts optional[context]`,
	Run: func(cmd *cobra.Command, args []string) {
		context := config.Config().GetAppConfig().GetContext()
		if len(args) > 0 && len(args[0]) > 0 {
			context = args[0]
		}
		config.Config().PrintContext(context)

	},
}

func init() {
	rootCmd.AddCommand(context)
	context.AddCommand(contextClear)
	context.AddCommand(contextCopy)
	context.AddCommand(contextDelete)
	context.AddCommand(contextList)
	context.AddCommand(contextNew)
	context.AddCommand(contextSet)
	context.AddCommand(contextShow)
}
