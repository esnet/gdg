package containers

import (
	"log/slog"
	"os"
	"testing"
)

func TestGetGrafanaVersion(t *testing.T) {
	grafanaVersion := os.Getenv(defaultGrafanaVersionEnv)
	testGrafanaVersion := "12.3.0-ubuntu"
	os.Setenv(defaultGrafanaVersionEnv, testGrafanaVersion)
	defer func() {
		envErr := os.Setenv(defaultGrafanaVersionEnv, grafanaVersion)
		if envErr != nil {
			slog.Error("Error setting env vars", "envErr", envErr)
		}
	}()

	tests := []struct {
		name string
		want string
	}{
		{
			name: "Basic Test",
			want: testGrafanaVersion,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetGrafanaVersion(); got != tt.want {
				t.Errorf("GetGrafanaVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
