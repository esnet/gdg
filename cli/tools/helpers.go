package tools

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/bep/simplecobra"
	"github.com/esnet/gdg/cli/domain"
	"github.com/esnet/gdg/pkg/encode"
	"github.com/spf13/cobra"
)

func newHelpers() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "helpers",
		Short: "Config Helpers",
		Long:  "Config Helpers",
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Aliases = []string{"h"}
		},
		CommandsList: []simplecobra.Commander{
			newFolderHelper(),
			newCipherHelper(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newCipherHelper() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "cipher",
		Short: "Cipher Helpers",
		Long:  "Cipher Helpers",
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Aliases = []string{"c", "ciphers"}
			cmd.PersistentFlags().StringP("file", "f", "", "file to encode/decode")
			cmd.PersistentFlags().StringP("value", "", "", "value to encode/decode")
		},
		CommandsList: []simplecobra.Commander{
			newCipherEncode(),
			newCipherDecode(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newFolderHelper() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "folder",
		Short: "Config Helpers",
		Long:  "Config Helpers",
		WithCFunc: func(cmd *cobra.Command, r *domain.RootCommand) {
			cmd.Aliases = []string{"f", "folders"}
		},
		CommandsList: []simplecobra.Commander{
			newFolderEncode(),
			newFolderDecode(),
		},
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			return cd.CobraCommand.Help()
		},
	}
}

func newFolderEncode() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "encode",
		Short: "encode folder name as regex",
		Long:  "encode folder name as regex",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
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
	return &domain.SimpleCommand{
		NameP: "decode",
		Short: "decode folder name from regex",
		Long:  "decode folder name from regex",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("requires the following parameters to be specified: \"Folder Name\"")
			}
			folderName := args[0]
			result := encode.EncodePath(encode.DecodeEscapeSpecialChars, folderName)
			slog.Info("Decoded result", "output", result)
			return nil
		},
	}
}

func newCipherEncode() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "encode",
		Short: "apply cipher to string",
		Long:  "apply cipher to string",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			fileName, _ := cd.CobraCommand.Flags().GetString("file")
			value, _ := cd.CobraCommand.Flags().GetString("value")
			if fileName != "" && value != "" {
				log.Fatal("either a value or a file must be specified, not both")
			}
			if value != "" {
				result := rootCmd.GrafanaSvc().EncodeValue(value)
				slog.Info("Encoded result:")
				fmt.Println(result)
			} else {
				data, err := os.ReadFile(fileName) // #nosec G304
				if err != nil {
					log.Fatal("Error reading file", "file", fileName, "err", err)
				}

				result := rootCmd.GrafanaSvc().EncodeValue(string(data))
				if result != "" {
					err = os.WriteFile(fileName, []byte(result), 0o600)
					if err != nil {
						log.Fatal("Error writing file", "file", fileName, "err", err)
					} else {
						slog.Info("File has been encrypted", "file", fileName)
					}
				}
			}

			return nil
		},
	}
}

func newCipherDecode() simplecobra.Commander {
	return &domain.SimpleCommand{
		NameP: "decode",
		Short: "decode string using cipher plugin",
		Long:  "decode string using cipher plugin",
		RunFunc: func(ctx context.Context, cd *simplecobra.Commandeer, rootCmd *domain.RootCommand, args []string) error {
			fileName, _ := cd.CobraCommand.Flags().GetString("file")
			value, _ := cd.CobraCommand.Flags().GetString("value")
			if fileName != "" && value != "" {
				log.Fatal("either a value or a file must be specified, not both")
			}
			if value != "" {
				result := rootCmd.GrafanaSvc().DecodeValue(value)
				slog.Info("Decoded result")
				fmt.Println(result)
			} else {
				data, err := os.ReadFile(fileName) // #nosec G304
				if err != nil {
					log.Fatal("Error reading file", "file", fileName, "err", err)
				}

				result := rootCmd.GrafanaSvc().DecodeValue(string(data))
				if result != "" {
					err = os.WriteFile(fileName, []byte(result), 0o600)
					if err != nil {
						log.Fatal("Error writing file", "file", fileName, "err", err)
					} else {
						slog.Info("File has been decrypted", "file", fileName)
					}
				}
			}
			return nil
		},
	}
}
