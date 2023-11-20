package config

import (
	"embed"
	"log/slog"
)

//go:embed *
var Assets embed.FS

func GetFile(name string) (string, error) {
	data, err := Assets.ReadFile(name)
	if err != nil {
		slog.Info("unable to find load default configuration", "err", err)
		return "", err
	}
	return string(data), nil
}
