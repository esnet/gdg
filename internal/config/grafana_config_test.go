package config

import (
	"testing"

	"github.com/esnet/gdg/internal/config/domain"
)

func TestGrafanaConfig(t *testing.T) {
	config := domain.GrafanaConfig{
		URL: "  http://localhost  ",
	}
	expected := "http://localhost/"

	if expected != config.GetURL() {
		t.Errorf("expected %s, got %s", expected, config.GetURL())
	}
}
