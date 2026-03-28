package config

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/esnet/gdg/internal/config/config_domain"
)

const (
	cloudKindKey   = "kind"
	cloudKindCloud = "cloud"
	sslEnabledKey  = "ssl_enabled"
)

type cloudProvider string

const (
	providerCustom cloudProvider = "custom"
	providerAWS    cloudProvider = "s3"
	providerGCS    cloudProvider = "gs"
	providerAzure  cloudProvider = "azblob"
)

var providerDocURLs = map[cloudProvider]string{
	providerAWS:   "https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html",
	providerGCS:   "https://cloud.google.com/go/storage (auth: https://cloud.google.com/docs/authentication/application-default-credentials)",
	providerAzure: "https://learn.microsoft.com/en-us/azure/storage/common/storage-auth",
}

// ListS3Configs returns all storage_engine entries from the config.
func ListS3Configs(app *config_domain.GDGAppConfiguration) map[string]map[string]string {
	if app.StorageEngine == nil {
		return make(map[string]map[string]string)
	}
	return app.StorageEngine
}

// DeleteS3Config removes a named storage engine from the config and cleans up its
// credentials file from the current context's secure location. It warns (but does
// not fail) if the engine is still referenced by an existing context.
func DeleteS3Config(app *config_domain.GDGAppConfiguration, name string) {
	if _, exists := app.StorageEngine[name]; !exists {
		log.Fatalf("storage engine %q not found", name)
	}

	for ctxName, ctx := range app.GetContexts() {
		if ctx.Storage == name {
			ctx.Storage = ""
			slog.Info("cleared storage assignment from context", "context", ctxName)
		}
	}

	delete(app.StorageEngine, name)

	if err := app.SaveToDisk(false); err != nil {
		log.Fatal("failed to save configuration after delete")
	}

	// Remove the credentials file if it exists
	securePath := app.GetDefaultGrafanaConfig().SecureLocation()
	credFile := filepath.Join(
		securePath,
		fmt.Sprintf("%s_%s.yaml", config_domain.CloudAuthPrefix, name),
	)
	if _, statErr := os.Stat(credFile); statErr == nil {
		if removeErr := os.Remove(credFile); removeErr != nil {
			slog.Warn("failed to remove credentials file", "file", credFile)
		} else {
			slog.Info("credentials file removed", "file", credFile)
		}
	}

	slog.Info("storage engine deleted", "label", name)
}
