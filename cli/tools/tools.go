package tools

import (
	"context"
	"slices"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/domain"
	"github.com/spf13/cobra"
)

// noLoginGroups lists the top-level tools subcommands whose leaf commands
// operate purely on local config and never need a live Grafana connection.
var noLoginGroups = map[string]bool{
	"contexts": true,
	"helpers":  true,
	"plugins":  true,
}

// needsLogin walks runner's Cobra parent chain from the leaf up to (but not
// including) the "tools" command.  If any command in that range is in
// noLoginGroups, no Grafana login is required.
func needsLogin(runner *simplecobra.Commandeer) bool {
	cmd := runner.CobraCommand
	for cmd != nil && cmd.Name() != "tools" {
		if noLoginGroups[cmd.Name()] {
			return false
		}
		cmd = cmd.Parent()
	}
	return true
}

func getBasicRoles() []string {
	return []string{"admin", "editor", "viewer"}
}

func validBasicRole(role string) bool {
	return slices.Contains(getBasicRoles(), role)
}

func NewToolsCommand() simplecobra.Commander {
	description := "A collection of tools to manage a grafana instance"
	return &domain.SimpleCommand{
		NameP: "tools",
		Short: description,
		Long:  description,
		CommandsList: []simplecobra.Commander{
			newContextCmd(),
			newDevelCmd(),
			newUserCommand(),
			newAuthCmd(),
			newOrgCommand(),
			newHelpers(),
			newPluginsCmd(),
		},
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Aliases = []string{"t"}
		},
		InitCFunc: func(cd *simplecobra.Commandeer, runner *simplecobra.Commandeer, r *domain.RootCommand) error {
			configOverride, _ := cd.CobraCommand.Flags().GetString("config")
			contextOverride, _ := cd.CobraCommand.Flags().GetString("context")
			r.LoadConfig(configOverride, contextOverride)
			if needsLogin(runner) {
				r.GrafanaSvc().Login()
			}
			return nil
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}
