package test_tooling

import (
	"context"
	"fmt"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/pkg/test_tooling/containers"
	"log"
	"log/slog"
	"os"
)

func SetupCloudFunction(params []string) (context.Context, context.CancelFunc, service.GrafanaService, error) {
	errorFunc := func(err error) (context.Context, context.CancelFunc, service.GrafanaService, error) {
		return nil, nil, nil, err
	}
	_ = os.Setenv(service.InitBucket, "true")
	bucketName := params[1]
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
	var m = map[string]string{
		service.InitBucket: "true",
		service.CloudType:  params[0],
		service.Prefix:     "dummy",
		service.AccessId:   "test",
		service.SecretKey:  "secretsss",
		service.BucketName: bucketName,
		service.Kind:       "cloud",
		service.Custom:     "true",
		service.Endpoint:   minioHost,
		service.SSLEnabled: "false",
	}

	cfgObj := config.Config().GetGDGConfig()
	defaultCfg := config.Config().GetDefaultGrafanaConfig()
	defaultCfg.Storage = "test"
	cfgObj.StorageEngine["test"] = m
	apiClient := service.NewApiService("dummy")

	ctx := context.Background()
	ctx = context.WithValue(ctx, service.StorageContext, m)
	configMap := map[string]string{}
	for key, value := range m {
		configMap[key] = fmt.Sprintf("%v", value)
	}

	s, err := service.NewCloudStorage(ctx)
	if err != nil {
		log.Fatalf("Could not instantiate cloud storage for type: %s", params[0])
	}
	dash := apiClient.(*service.DashNGoImpl)
	dash.SetStorage(s)

	return ctx, cancel, apiClient, nil
}
