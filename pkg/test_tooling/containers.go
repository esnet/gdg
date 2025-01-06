package test_tooling

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/esnet/gdg/internal/storage"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling/containers"
)

type CloudTestOpt func(m *map[string]string)

func SetBucketName(bucketName string) CloudTestOpt {
	return func(m *map[string]string) {
		(*m)[storage.BucketName] = bucketName
	}
}

func SetCloudType(cloudType string) CloudTestOpt {
	return func(m *map[string]string) {
		(*m)[storage.CloudType] = cloudType
	}
}

func SetPrefix(prefix string) CloudTestOpt {
	return func(m *map[string]string) {
		(*m)[storage.Prefix] = prefix
	}
}

func SetupCloudFunctionOpt(opts ...CloudTestOpt) (context.Context, context.CancelFunc, service.GrafanaService, error) {
	errorFunc := func(err error) (context.Context, context.CancelFunc, service.GrafanaService, error) {
		return nil, nil, nil, err
	}
	_ = os.Setenv(storage.InitBucket, "true")
	container, cancel := containers.BootstrapCloudStorage("", "")
	wwwPort, err := container.PortEndpoint(context.Background(), "9001", "")
	if err != nil {
		return errorFunc(err)
	}
	actualPort, err := container.Endpoint(context.Background(), "")
	if err != nil {
		return errorFunc(err)
	}
	minioHost := fmt.Sprintf("http://%s", actualPort)
	slog.Info("Minio container is up and running", slog.Any("hostname", fmt.Sprintf("http://%s", wwwPort)))
	m := map[string]string{
		storage.InitBucket: "true",
		storage.CloudType:  "cloud",
		storage.Prefix:     "dummy",
		storage.AccessId:   "test",
		storage.SecretKey:  "secretsss",
		storage.BucketName: "testing",
		storage.Endpoint:   minioHost,
	}
	for _, opt := range opts {
		opt(&m)
	}

	cfgObj := config.Config().GetGDGConfig()
	defaultCfg := config.Config().GetDefaultGrafanaConfig()
	defaultCfg.Storage = "test"
	cfgObj.StorageEngine["test"] = m

	ctx := context.Background()
	ctx = context.WithValue(ctx, storage.Context, m)

	s, err := storage.NewCloudStorage(ctx)
	if err != nil {
		log.Fatalf("Could not instantiate cloud storage for type: %s", m[storage.CloudType])
	}

	apiClient := service.NewTestApiService(s, func() *config.Configuration {
		return config.Config()
	})

	return ctx, cancel, apiClient, nil
}
