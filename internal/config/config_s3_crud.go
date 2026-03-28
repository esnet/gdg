package config

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"charm.land/huh/v2"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/cipher"
	"github.com/esnet/gdg/internal/adapter/plugins/secure/noop"
	"github.com/esnet/gdg/internal/adapter/storage"
	"github.com/esnet/gdg/internal/config/config_domain"
	"github.com/esnet/gdg/internal/ports"
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

// NewCustomS3Config runs a TUI wizard to configure a cloud storage engine.
// For managed providers (AWS S3, GCS, Azure) it prints the relevant docs URL and returns.
// For custom S3-compatible providers it collects config, writes credentials to the
// current context's secure location (matching the path GetCloudAuthLocation resolves),
// persists non-sensitive config to gdg.yml, and optionally assigns the storage engine
// to the active context.
func NewCustomS3Config(app *config_domain.GDGAppConfiguration) {
	// Step 1 — provider selection
	var provider string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Cloud storage provider").
				Description("For AWS S3, GCS, and Azure, GDG delegates auth to the provider SDK.\nOnly custom S3-compatible endpoints (Minio, Ceph, etc.) are configured here.").
				Options(
					huh.NewOption("Custom S3-compatible (Minio, Ceph, …)", string(providerCustom)),
					huh.NewOption("AWS S3", string(providerAWS)),
					huh.NewOption("Google Cloud Storage (GCS)", string(providerGCS)),
					huh.NewOption("Azure Blob Storage", string(providerAzure)),
				).
				Value(&provider),
		),
	).WithShowHelp(false).WithShowErrors(false).Run()
	if err != nil {
		slog.Warn("storage configuration cancelled — no storage engine added")
		return
	}

	if cp := cloudProvider(provider); cp != providerCustom {
		fmt.Printf("\nFor %s, authentication is handled by the provider SDK — not configured here.\n", cp)
		fmt.Printf("Documentation: %s\n\n", providerDocURLs[cp])
		fmt.Println("Once credentials are in place, add a storage_engine entry to gdg.yml:")
		fmt.Printf("  cloud_type: %s\n  bucket_name: <your-bucket>\n\n", cp)
		slog.Info("storage engine not configured here — add gdg.yml entry manually", "provider", string(cp))
		return
	}

	// Step 2 — custom S3 fields
	var (
		label      string
		endpoint   string
		bucket     string
		region     string
		accessID   string
		secretKey  string
		prefix     string
		initBucket bool
		sslEnabled bool
	)

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Storage engine label").
				Description("Unique key for this config in gdg.yml (e.g. my-minio)").
				Value(&label),
			huh.NewInput().
				Title("Endpoint URL").
				Description("Full URL of the S3-compatible endpoint (e.g. http://localhost:9000)").
				Value(&endpoint),
			huh.NewInput().
				Title("Bucket name").
				Value(&bucket),
			huh.NewInput().
				Title("Region").
				Description("AWS region or equivalent (default: us-east-1)").
				Value(&region),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Access Key ID").
				Value(&accessID),
			huh.NewInput().
				Title("Secret Access Key").
				EchoMode(huh.EchoModePassword).
				Value(&secretKey),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Path prefix (optional)").
				Description("Prefix applied to all object paths within the bucket").
				Value(&prefix),
			huh.NewConfirm().
				Title("Auto-create bucket if it does not exist?").
				Value(&initBucket),
			huh.NewConfirm().
				Title("Enable SSL?").
				Value(&sslEnabled),
		),
	).Run()
	if err != nil {
		log.Fatalf("storage configuration cancelled: %v", err)
	}

	if label == "" || endpoint == "" || bucket == "" {
		log.Fatal("label, endpoint, and bucket name are required")
	}

	if app.StorageEngine == nil {
		app.StorageEngine = make(map[string]map[string]string)
	}
	if _, exists := app.StorageEngine[label]; exists {
		log.Fatalf("storage engine %q already exists; delete it first or choose a different label", label)
	}

	// Initialise the cipher encoder — mirrors CreateNewContext.
	// If no plugin is configured, NoOpEncoder passes values through unchanged.
	var encoder ports.CipherEncoder
	if !app.PluginConfig.Disabled && app.PluginConfig.CipherPlugin != nil {
		var encErr error
		encoder, encErr = cipher.NewPluginCipherEncoder(app.PluginConfig.CipherPlugin, app.SecureConfig)
		if encErr != nil {
			log.Fatalf("Failed to load cipher plugin: %v", encErr)
		}
	} else {
		encoder = noop.NoOpEncoder{}
	}

	// Encode the secret key before persisting. The cloud storage adapter already
	// calls encoder.DecodeValue on SecretKey at read time, so this round-trips correctly.
	encodedSecret, encErr := encoder.EncodeValue(secretKey)
	if encErr != nil {
		log.Fatalf("failed to encode secret key: %v", encErr)
	}

	// Write credentials to the current context's secure location.
	// The path mirrors what GetCloudAuthLocation() resolves to, so the existing
	// cloud storage adapter's GetCloudAuth() picks them up automatically.
	grafanaCfg := app.GetDefaultGrafanaConfig()
	securePath := grafanaCfg.SecureLocation()
	if mkErr := os.MkdirAll(securePath, 0o750); mkErr != nil {
		log.Fatalf("unable to create secure directory %s: %v", securePath, mkErr)
	}

	credFile := filepath.Join(
		securePath,
		fmt.Sprintf("%s_%s.yaml", config_domain.CloudAuthPrefix, label),
	)
	creds := map[string]string{
		storage.AccessId:  accessID,
		storage.SecretKey: encodedSecret,
	}
	if writeErr := writeSecureFileData(creds, credFile); writeErr != nil {
		log.Fatalf("unable to write credentials file %s: %v", credFile, writeErr)
	}

	// Persist non-sensitive storage engine config to gdg.yml.
	// Credentials are intentionally omitted here — they live in the secure file above.
	app.StorageEngine[label] = map[string]string{
		cloudKindKey:       cloudKindCloud,
		storage.CloudType:  storage.Custom,
		storage.Endpoint:   endpoint,
		storage.BucketName: bucket,
		storage.Region:     region,
		storage.Prefix:     prefix,
		storage.InitBucket: fmt.Sprintf("%t", initBucket),
		sslEnabledKey:      fmt.Sprintf("%t", sslEnabled),
	}

	// Step 3 — optionally assign to the active context
	var assignToCtx bool
	activeCtx := app.GetContext()
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("Assign this storage engine to the active context (%q)?", activeCtx)).
				Value(&assignToCtx),
		),
	).WithShowHelp(false).WithShowErrors(false).Run()
	if err != nil {
		log.Fatalf("context assignment cancelled: %v", err)
	}

	if assignToCtx {
		grafanaCfg.Storage = label
	}

	if saveErr := app.SaveToDisk(false); saveErr != nil {
		log.Fatal("failed to save configuration to disk")
	}

	slog.Info("Storage engine created", "label", label, "endpoint", endpoint, "context_assigned", assignToCtx)
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
