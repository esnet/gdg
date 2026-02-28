package main

import (
	"log"
	"os"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/internal/adapter/grafana/api"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/cipher"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/noop"
	"github.com/esnet/gdg/internal/adapter/storage"
	configDomain "github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/ports"
	"github.com/esnet/gdg/pkg/cli_helper"
)

func main() {
	// peek at flags before cobra runs
	configPath, contextOverride := cli_helper.ParseConfigContextParams()

	// build adapters now that we know the config path
	rootSvc := cli.NewRootService()
	cfg := rootSvc.LoadConfig(configPath, contextOverride)
	grafanaSvc := buildGrafanaService(cfg)
	rootSvc.SetService(grafanaSvc)

	// hand off to CLI with everything wired
	err := cli.Execute(rootSvc, os.Args[1:])
	if err != nil {
		log.Fatalf("Error: %s", err)
	}
}

// buildGrafanaService constructs and returns a GrafanaService instance from the given application configuration.
// It configures a cipher encoder based on plugin settings, resolves the storage engine from cloud configuration,
// and initializes the service. The function will terminate the process if the storage engine cannot be configured.
func buildGrafanaService(cfg *configDomain.GDGAppConfiguration) ports.GrafanaService {
	var encoder ports.CipherEncoder
	if !cfg.PluginConfig.Disabled && cfg.PluginConfig.CipherPlugin != nil {
		encoder = cipher.NewPluginCipherEncoder(cfg.PluginConfig.CipherPlugin, cfg.SecureConfig)
	} else {
		encoder = noop.NoOpEncoder{}
	}

	storageType, appData := cfg.GetCloudConfiguration(cfg.GetDefaultGrafanaConfig().Storage)

	storageEngine, err := storage.NewStorageFromConfig(storageType, appData, encoder)
	if err != nil {
		log.Fatal("Unable to configure a valid storage engine, %w", err)
	}
	// TODO: wire NewExtendedApi in main instead of relying on config values.
	grafanaSvc := api.NewDashNGo(cfg, encoder, storageEngine)
	return grafanaSvc
}
