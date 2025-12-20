package test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/joho/godotenv"
)

const (
	defaultGrafanaVersion = "11.6.0-ubuntu"
	basicAuth             = "basicAuth"
	testDebug             = "TEST_DEBUG"
	developerEnv          = "DEVELOPER"
	DefaultRetryAttempts  = 3
)

func TestMain(m *testing.M) {
	gofakeit.Seed(time.Now().Unix()) // If 0 will use crypto/rand to generate a number
	err := path.FixTestDir("test", "..")
	if err != nil {
		panic(err)
	}

	err = godotenv.Load(".env")
	// set global log level
	if os.Getenv(testDebug) == "1" {
		slog.SetLogLoggerLevel(slog.LevelDebug) // Set global log level to Debug
	}

	developer := getEnvDefault(developerEnv, test_tooling.FeatureDisabled)
	version := getEnvDefault(test_tooling.GrafanaTestVersionEnv, defaultGrafanaVersion)
	// When developer is enabled both token and basic auth tests are executed.
	if developer == test_tooling.FeatureEnabled {
		for _, tokenVal := range []string{"0", "1"} {
			os.Setenv(test_tooling.EnableTokenTestsEnv, tokenVal)
			runTests(version, tokenVal, m)
		}
	} else {
		tokenVal := getEnvDefault(test_tooling.EnableTokenTestsEnv, "0")
		runTests(version, tokenVal, m)
	}
}

// runTests executes the test suite with a specified Grafana version and authentication mode,
// setting environment variables for the tests and panicking if any test fails.
func runTests(version, token string, m *testing.M) {
	var tokenDesc string
	if token == test_tooling.FeatureEnabled {
		tokenDesc = token
	} else {
		tokenDesc = basicAuth
	}
	slog.Info("Running Test suit for",
		slog.Any("grafanaVersion", version),
		slog.Any("AuthMode", tokenDesc))

	os.Setenv(test_tooling.GrafanaTestVersionEnv, version)
	os.Setenv(test_tooling.EnableTokenTestsEnv, token)

	exitVal := m.Run()
	if exitVal != 0 {
		panic("Failed to run test with token value of: " + tokenDesc)
	}
}

// getGrafanaVersion extracts the major Grafana version from a tag string, returning 0 if parsing fails.
func getGrafanaVersion(tag string) int {
	parts := strings.Split(tag, ":")
	if len(parts) < 2 {
		return 0
	}

	version := parts[len(parts)-1]
	parts = strings.Split(version, ".")

	ver, err := strconv.Atoi(parts[0])
	if err != nil {
		slog.Error("failed to convert string version to a numeric value", slog.Any("err", err))
	}
	return ver
}

// diffStruct compares two values of any type and logs a mismatch if they differ.
func diffStruct[T any](a, expected T) bool {
	if !cmp.Equal(a, expected) {
		slog.Error("diffStruct[...] mismatch", "type", reflect.TypeOf(a))
		fmt.Println(cmp.Diff(a, expected))
		return false
	}
	return true
}

// TestGrafanaVersion verifies that getGrafanaVersion correctly parses major version numbers from image tags.
func TestGrafanaVersion(t *testing.T) {
	image := "grafana/grafana:10.2.8"
	expectedVal := 10
	assert.Equal(t, expectedVal, getGrafanaVersion(image))
	image = "grafana/grafana:11.2.2"
	expectedVal = 11
	assert.Equal(t, 11, getGrafanaVersion(image))
	image = "grafana/grafana"
	assert.Equal(t, 0, getGrafanaVersion(image))
}

// getEnvDefault returns the environment variable value for key or defaultValue if unset or empty.
func getEnvDefault(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val
}

type RetryFunc func() error

// Retry attempts to execute f up to retryAttempts times, returning nil on success or the last error.
// It stops early if ctx is cancelled and returns ctx.Err().
func Retry(ctx context.Context, retryAttempts int, f RetryFunc) error {
	var err error
	for i := 0; i < retryAttempts; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = f()
			if err == nil {
				return nil
			}
		}
	}
	return err
}
