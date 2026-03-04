package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/domain"
	"github.com/esnet/gdg/internal/adapter/storage"
	"github.com/esnet/gdg/internal/config"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func newS3Cmd() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "s3",
		Short: "Manage custom S3-compatible storage engine configurations",
		Long:  "Create, list, and delete custom S3-compatible storage engine configs. For AWS S3, GCS, or Azure, the wizard prints the relevant documentation URL instead.",
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Aliases = []string{"storage"}
		},
		CommandsList: []simplecobra.Commander{
			newS3NewCmd(),
			newS3ListCmd(),
			newS3DeleteCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, r *domain.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newS3NewCmd() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "new",
		Short: "Launch the TUI wizard to create a new cloud storage engine config",
		Long:  "Interactively configure a custom S3-compatible storage engine (Minio, Ceph, etc.). For AWS S3, GCS, or Azure, prints the relevant auth documentation URL.",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, r *domain.RootCommand, args []string) error {
			config.NewCustomS3Config(r.ConfigSvc())
			return nil
		},
	}
}

func newS3ListCmd() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "list",
		Short: "List all configured storage engines",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, r *domain.RootCommand, args []string) error {
			engines := config.ListS3Configs(r.ConfigSvc())
			if len(engines) == 0 {
				fmt.Println("No storage engines configured.")
				return nil
			}

			// Build a reverse index: storage label → context names that reference it
			assignedTo := make(map[string][]string)
			for ctxName, ctx := range r.ConfigSvc().GetContexts() {
				if ctx.Storage != "" {
					assignedTo[ctx.Storage] = append(assignedTo[ctx.Storage], ctxName)
				}
			}

			r.TableObj.AppendHeader(table.Row{
				"label", "cloud_type", "endpoint", "bucket", "region", "prefix", "init_bucket", "ssl", "assigned_contexts",
			})

			for label, cfg := range engines {
				assigned := "-"
				if refs := assignedTo[label]; len(refs) > 0 {
					assigned = strings.Join(refs, ", ")
				}
				r.TableObj.AppendRow(table.Row{
					label,
					cfg[storage.CloudType],
					cfg[storage.Endpoint],
					cfg[storage.BucketName],
					cfg[storage.Region],
					cfg[storage.Prefix],
					cfg[storage.InitBucket],
					cfg["ssl_enabled"],
					assigned,
				})
			}

			r.Render(cd.CobraCommand, engines)
			return nil
		},
	}
}

func newS3DeleteCmd() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "delete",
		Short: "delete <label>",
		Long:  "Delete a named storage engine config and its credentials file from the secure location.",
		InitCFunc: func(cd *simplecobra.Commandeer, r *domain.RootCommand) error {
			cd.CobraCommand.Aliases = []string{"del"}
			return nil
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, r *domain.RootCommand, args []string) error {
			if len(args) < 1 {
				return cd.CobraCommand.Help()
			}
			config.DeleteS3Config(r.ConfigSvc(), args[0])
			return nil
		},
	}
}
