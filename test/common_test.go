package test

import (
	"github.com/esnet/gdg/pkg/test_tooling"
	"github.com/esnet/gdg/pkg/test_tooling/path"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	err := path.FixTestDir("test", "..")
	if err != nil {
		panic(err)
	}

	err = godotenv.Load(".env")
	//set global log level
	slog.SetLogLoggerLevel(slog.LevelDebug) // Set global log level to Debug
	grafanaTestVersions := []string{"10.2.3-ubuntu", "11.1.5-ubuntu"}
	testModes := []string{"basicAuth", "token"}
	if os.Getenv("DEVELOPER") == "1" {
		slog.Debug("Limiting to single testMode and grafana version", slog.Any("grafanaVersion", grafanaTestVersions[1]), slog.String("testMode", testModes[0]))
		grafanaTestVersions = grafanaTestVersions[1:]
		testModes = testModes[0:1]
	}

	for _, version := range grafanaTestVersions {
		for _, i := range testModes {
			os.Setenv(test_tooling.GrafanaTestVersionEnv, version)
			if i == "token" {
				os.Setenv(test_tooling.EnableTokenTestsEnv, "1")
			} else {
				os.Setenv(test_tooling.EnableTokenTestsEnv, "0")
			}
			slog.Info("Running Test suit for",
				slog.Any("grafanaVersion", version),
				slog.Any("AuthMode", i))
			exitVal := m.Run()
			if exitVal != 0 {
				panic("Failed to run test with token value of: " + i)
			}
		}
	}
}
