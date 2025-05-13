package test

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/joho/godotenv"
)

const (
	minimumNestedFoldersVersion = 11
	defaultGrafanaVersion       = "11.1.5-ubuntu"
	basicAuth                   = "basicAuth"
	developerEnv                = "DEVELOPER"
)

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

func TestMain(m *testing.M) {
	gofakeit.Seed(time.Now().Unix()) // If 0 will use crypto/rand to generate a number
	err := path.FixTestDir("test", "..")
	if err != nil {
		panic(err)
	}

	err = godotenv.Load(".env")
	// set global log level
	slog.SetLogLoggerLevel(slog.LevelDebug) // Set global log level to Debug

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

func getEnvDefault(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	return val
}
