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
	DisableEnterpriseTest    = "ENTERPRISE_DISABLED"
	DefaultCloudUser         = "test"
	DefaultCloudPass         = "secretsss"
	minioCurrentTag          = "RELEASE.2025-09-07T16-13-09Z"
)

// BootstrapCloudStorage starts a S3 container for cloud storage testing.
// It accepts optional username and password; defaults are used if empty.
// Returns the testcontainers.Container and a cancel function to terminate it.
func BootstrapCloudStorage(username, password string) (testcontainers.Container, context.CancelFunc) {
	if username == "" || password == "" {
		username = DefaultCloudUser
		password = DefaultCloudPass
	}

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        fmt.Sprintf("minio/minio:%s", minioCurrentTag),
		Cmd:          []string{"server", "start", "--console-address", ":9001"},
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

// SetupGrafanaLicense loads the enterprise license from ENTERPRISE_LICENSE env var,
// stores it under GF_ENTERPRISE_LICENSE_TEXT in props, and errors if not set.
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

// GetGrafanaVersion returns the Grafana version used for tests.
// It reads GRAFANA_TEST_VERSION env variable; if unset, defaults to "11.1.4-ubuntu".
func GetGrafanaVersion() string {
	version := os.Getenv(defaultGrafanaVersionEnv)
	if version == "" {
		version = defaultGrafanaVersion
	}
	return version
}

// SetupGrafanaContainer starts a Grafana test container with default env vars,
// merges additionalEnvProps, retries up to 3 times, and returns the container
// and a cancel function to terminate it.
func SetupGrafanaContainer(additionalEnvProps map[string]string, version, imageSuffix string) (testcontainers.Container, func()) {
	retry := func() (testcontainers.Container, func(), error) {
		defaultProps := DefaultGrafanaEnv()
		version = GetGrafanaVersion()
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
	for range 3 {
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
