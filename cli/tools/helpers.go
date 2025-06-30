package tools

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/support"
	"github.com/esnet/gdg/internal/tools/encode"
	"github.com/spf13/cobra"
)

func newHelpers() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "helpers",
		Short: "Config Helpers",
		Long:  "Config Helpers",
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"h"}
		},
		CommandsList: []simplecobra.Commander{
			newFolderHelper(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newFolderHelper() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "folder",
		Short: "Config Helpers",
		Long:  "Config Helpers",
		WithCFunc: func(cmd *cobra.Command, r *support.RootCommand) {
			cmd.Aliases = []string{"f", "folders"}
		},
		CommandsList: []simplecobra.Commander{
			newFolderEncode(),
			newFolderDecode(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newFolderEncode() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "encode",
		Short: "encode folder name as regex",
		Long:  "encode folder name as regex",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("requires the following parameters to be specified: \"Folder Name\"")
			}
			folderName := args[0]
			result := encode.EncodePath(encode.EncodeEscapeSpecialChars, folderName)
			slog.Info("Encoded result", "output", result)
			return nil
		},
	}
}

func newFolderDecode() simplecobra.Commander {
	return &support.SimpleCommand{
		NameP: "decode",
		Short: "decode folder name from regex",
		Long:  "decode folder name from regex",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *support.RootCommand, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("requires the following parameters to be specified: \"Folder Name\"")
			}
			folderName := args[0]
			// result := encode.DecodeEscapeSpecialChars(folderName)
			result := encode.EncodePath(encode.DecodeEscapeSpecialChars, folderName)
			slog.Info("Decoded result", "output", result)
			return nil
		},
	}
}
