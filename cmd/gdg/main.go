package main

import (
	"log"
	"os"

	"github.com/esnet/gdg/cli"
	"github.com/esnet/gdg/internal/adapter/grafana/api"
	"github.com/esnet/gdg/internal/adapter/grafana/extended"
	"github.com/esnet/gdg/internal/adapter/grafana/resources"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/cipher"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/noop"
	"github.com/esnet/gdg/internal/adapter/storage"
	configDomain "github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/ports/outbound"
)

func main() {
	// Register the service factory so config-dependent commands can build the
	// GrafanaService on first use.  Commands that don't need config (version,
	// default-config, help) run without ever touching the factory.
	rootSvc := cli.NewRootService()
	rootSvc.SetServiceFactory(buildGrafanaService)

	if err := cli.Execute(rootSvc, os.Args[1:]); err != nil {
		log.Fatalf("Error: %s", err)
	}
}

// buildGrafanaService constructs and returns a GrafanaService instance from the given application configuration.
// It configures a cipher encoder based on plugin settings, resolves the storage engine from cloud configuration,
// and initializes the service. The function will terminate the process if the storage engine cannot be configured.
func buildGrafanaService(cfg *configDomain.GDGAppConfiguration) outbound.GrafanaService {
	var encoder outbound.CipherEncoder
	if !cfg.PluginConfig.Disabled && cfg.PluginConfig.CipherPlugin != nil {
		var err error
		encoder, err = cipher.NewPluginCipherEncoder(cfg.PluginConfig.CipherPlugin, cfg.SecureConfig)
		if err != nil {
			log.Fatalf("Failed to load cipher plugin: %v", err)
		}
	} else {
		encoder = noop.NoOpEncoder{}
	}

	storageType, appData := cfg.GetCloudConfiguration(cfg.GetDefaultGrafanaConfig().Storage)

	storageEngine, err := storage.NewStorageFromConfig(storageType, appData, encoder)
	if err != nil {
		log.Fatal("Unable to configure a valid storage engine, %w", err)
	}
	extendedApi := extended.NewExtendedApi(cfg)
	grafanaSvc := api.NewDashNGo(cfg, encoder, storageEngine, extendedApi, resources.NewHelpers())
	return grafanaSvc
}
