package config

import "testing"

func TestGrafanaConfig(t *testing.T) {
	config := GrafanaConfig{
		URL: "  http://localhost  ",
	}
	expected := "http://localhost/"

	if expected != config.GetURL() {
		t.Errorf("expected %s, got %s", expected, config.GetURL())
	}
}
