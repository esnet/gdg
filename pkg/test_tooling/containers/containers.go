package containers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"maps"
	"os"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	defaultGrafanaVersion    = "11.1.4-ubuntu"
	defaultGrafanaVersionEnv = "GRAFANA_TEST_VERSION"
	EnterpriseLicenceKey     = "GF_ENTERPRISE_LICENSE_TEXT"
	EnterpriseLicenceKeyEnv  = "ENTERPRISE_LICENSE"
	DefaultCloudUser         = "test"
	DefaultCloudPass         = "secretsss"
)

func BootstrapCloudStorage(username, password string) (testcontainers.Container, context.CancelFunc) {
	if username == "" || password == "" {
		username = DefaultCloudUser
		password = DefaultCloudPass
	}

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "bitnami/minio:latest",
		ExposedPorts: []string{"9000/tcp", "9001/tcp"},
		Env:          map[string]string{"MINIO_ROOT_USER": username, "MINIO_ROOT_PASSWORD": password},
		WaitingFor:   wait.ForListeningPort("9000/tcp"),
	}
	minioC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}

	cancel := func() {
		if err := minioC.Terminate(ctx); err != nil {
			panic(err)
		} else {
			slog.Info("Minio container has been terminated")
		}
	}
	return minioC, cancel
}

func SetupGrafanaLicense(props *map[string]string) error {
	val := os.Getenv(EnterpriseLicenceKeyEnv)
	(*props)[EnterpriseLicenceKey] = val
	if val == "" {
		return errors.New("no valid enterprise license found")
	}
	return nil
}

func DefaultGrafanaEnv() map[string]string {
	return map[string]string{
		"GF_INSTALL_PLUGINS":         "grafana-googlesheets-datasource",
		"GF_AUTH_ANONYMOUS_ENABLED":  "true",
		"GF_SECURITY_ADMIN_PASSWORD": "admin", // This is a no-op right now, but we should trickle this up to
		// allow setting grafana admin credentials.
	}
}

func SetupGrafanaContainer(additionalEnvProps map[string]string, version, imageSuffix string) (testcontainers.Container, func()) {
	retry := func() (testcontainers.Container, func(), error) {
		defaultProps := DefaultGrafanaEnv()
		if version == "" {
			version = os.Getenv(defaultGrafanaVersionEnv)
			if version == "" {
				version = defaultGrafanaVersion
			}
		}
		// merge properties
		maps.Copy(defaultProps, additionalEnvProps)
		ctx := context.Background()
		req := testcontainers.ContainerRequest{
			Image:        fmt.Sprintf("grafana/grafana%s:%s", imageSuffix, version),
			ExposedPorts: []string{"3000/tcp"},
			Env:          defaultProps,
			WaitingFor:   wait.ForListeningPort("3000/tcp"),
		}
		grafanaC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("failed to retrieve valid container, %w", err)
		}

		cancel := func() {
			if err := grafanaC.Terminate(ctx); err != nil {
				slog.Warn("unable to terminate previous container", slog.Any("err", err))
			} else {
				slog.Info("Grafana Container has been terminated")
			}
		}

		return grafanaC, cancel, nil
	}

	// retry a few times just in case.
	for i := 0; i < 3; i++ {
		container, cancelFn, err := retry()
		if err == nil {
			return container, cancelFn
		} else {
			slog.Error(err.Error())
		}
	}

	log.Fatal("Unable to start container")
	return nil, nil
}
