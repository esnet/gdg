package backup

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

func newAlertingTimingsCommand() simplecobra.Commander {
	description := "Manage Alerting Mute Timings"
	return &support.SimpleCommand{
		NameP: "mute-timings",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"timings", "mute-timing", "timing", "t"}
		},
		CommandsList: []simplecobra.Commander{
			newListTimingsCmd(),
			newDownloadTimingsCmd(),
			newUploadTimingsCmd(),
			newClearTimingsCmd(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newListTimingsCmd() simplecobra.Commander {
	description := "List all alert timings for the given Organization"
	return &support.SimpleCommand{
		NameP: "list",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"l"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"name", "interval count", "settings"})
			slog.Info("Listing all alert timings",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))

			timingsList, err := rootCmd.GrafanaSvc().ListAlertTimings()
			if err != nil {
				log.Fatal("unable to list Orgs mute timings ", slog.Any("err", err))
			}
			if len(timingsList) == 0 {
				slog.Info("No mute timings found")
			} else {
				printAlertTimings(timingsList)
			}
			return nil
		},
	}
}

// getTimingsTableWriter creates a table.Writer configured for outputting timing summaries.
// It sets the output to os.Stdout, applies a light style, and appends a header row
// with columns "name" and "interval count", enabling auto-merge for the header.
func getTimingsTableWriter() table.Writer {
	writer := table.NewWriter()
	writer.SetOutputMirror(os.Stdout)
	writer.SetStyle(table.StyleLight)
	writer.AppendHeader(table.Row{"name", "interval count"}, table.RowConfig{AutoMerge: true})
	return writer
}

// printAlertTimings prints a summary table of mute timing intervals and detailed settings for each alert timing.
func printAlertTimings(timedIntervals []*models.MuteTimeInterval) {
	for _, link := range timedIntervals {
		writer := getTimingsTableWriter()
		writer.AppendRow(table.Row{link.Name, len(link.TimeIntervals)})
		writer.Render()
		var success bool
		twConfigs := table.NewWriter()
		twConfigs.SetOutputMirror(os.Stdout)
		twConfigs.SetStyle(table.StyleDouble)
		twConfigs.AppendHeader(table.Row{"Days", "Location", "Months", "Times", "Weekdays", "Years"})
		for _, timeInterval := range link.TimeIntervals {
			if timeInterval == nil {
				continue
			}
			success = true
			f := func(i any) string {
				raw, err := json.MarshalIndent(i, "", "  ")
				if err != nil {
					return "[]"
				}
				return string(raw)

			}
			twConfigs.AppendRow(table.Row{f(timeInterval.DaysOfMonth), timeInterval.Location,
				f(timeInterval.Months), f(timeInterval.Times), f(timeInterval.Weekdays), f(timeInterval.Years)})
		}
		if success {
			twConfigs.Render()
		}
	}
}

func newDownloadTimingsCmd() simplecobra.Commander {
	description := "Download all alert timings for the given Organization"
	return &support.SimpleCommand{
		NameP: "download",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"d"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"name", "interval count", "settings"})
			slog.Info("Download all alert timings",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))

			file, err := rootCmd.GrafanaSvc().DownloadAlertTimings()
			if err != nil {
				slog.Error("unable to download alert templates")
			} else {
				slog.Info("alert templates successfully downloaded", slog.Any("file", file))
			}
			return nil
		},
	}
}

func newUploadTimingsCmd() simplecobra.Commander {
	description := "Upload all alert timings for the given Organization"
	return &support.SimpleCommand{
		NameP: "upload",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"u"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			rootCmd.TableObj.AppendHeader(table.Row{"name"})
			slog.Info("Uploading all alert timings for context",
				slog.String("Organization", GetOrganizationName()),
				slog.String("context", GetContext()))

			files, err := rootCmd.GrafanaSvc().UploadAlertTimings()
			if err != nil {
				log.Fatal("unable to upload Orgs alerts timings", slog.Any("err", err))
			}
			for _, link := range files {
				rootCmd.TableObj.AppendRow(table.Row{link})
			}
			rootCmd.Render(cd.CobraCommand, files)
			return nil
		},
	}
}

func newClearTimingsCmd() simplecobra.Commander {
	description := "Delete all alert timings for the given Organization"
	return &support.SimpleCommand{
		NameP: "clear",
		Short: description,
		Long:  description,
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"c"}
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			slog.Info("Delete all alert timings")
			err := rootCmd.GrafanaSvc().ClearAlertTimings()
			if err != nil {
				log.Fatal("unable to deleting Orgs templates alerts", slog.Any("err", err))
			} else {
				slog.Info("alert timings successfully cleared")
			}
			return nil
		},
	}
}
