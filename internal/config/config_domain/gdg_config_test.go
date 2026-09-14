package config_domain

import (
	"bytes"
	"testing"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func TestPluginConfig_LookupEnabled(t *testing.T) {
	tests := []struct {
		name           string
		lookupDisabled bool
		want           bool
	}{
		{name: "enabled", lookupDisabled: false, want: true},
		{name: "lookup-specific disable", lookupDisabled: true, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &PluginConfig{
				Lookup: LookupConfig{Disabled: tt.lookupDisabled},
			}
			got := pc.LookupEnabled()
			if got != tt.want {
				t.Errorf("LookupEnabled() = %v, want %v (lookup.disabled=%v)",
					got, tt.want, tt.lookupDisabled)
			}
		})
	}
}

// TestPluginConfig_CipherEnabled verifies CipherEnabled follows the same
// "scoped disabled flag, not a global one" pattern as LookupEnabled: a nil
// CipherPlugin and an explicit CipherPlugin.Disabled=true both disable it.
func TestPluginConfig_CipherEnabled(t *testing.T) {
	tests := []struct {
		name   string
		plugin *CipherConfig
		want   bool
	}{
		{name: "nil cipher plugin", plugin: nil, want: false},
		{name: "configured and enabled", plugin: &CipherConfig{Disabled: false}, want: true},
		{name: "configured but disabled", plugin: &CipherConfig{Disabled: true}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := &PluginConfig{CipherPlugin: tt.plugin}
			got := pc.CipherEnabled()
			if got != tt.want {
				t.Errorf("CipherEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestLookupConfig_YAMLUnmarshal_SeparatesDisabledFromPlugins verifies that
// direct yaml.v3 decoding (used by the secure-config load path, e.g.
// config_loader.go's yaml.Unmarshal call) correctly splits the named
// "disabled" field from the inline map of named provider entries, rather
// than "disabled" leaking into Plugins as a bogus provider named "disabled".
func TestLookupConfig_YAMLUnmarshal_SeparatesDisabledFromPlugins(t *testing.T) {
	raw := []byte(`
disabled: true
gsm:
  url: https://example.com/lookup_gsm.wasm
  config:
    credentials: env:GOOGLE_APPLICATION_CREDENTIALS
vault:
  url: https://example.com/lookup_vault.wasm
`)

	var lc LookupConfig
	if err := yaml.Unmarshal(raw, &lc); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}

	if !lc.Disabled {
		t.Errorf("expected Disabled=true, got false")
	}
	if _, ok := lc.Plugins["disabled"]; ok {
		t.Errorf("\"disabled\" leaked into Plugins map, want it captured only by the Disabled field")
	}
	if len(lc.Plugins) != 2 {
		t.Fatalf("expected 2 provider entries in Plugins, got %d: %v", len(lc.Plugins), lc.Plugins)
	}
	gsm, ok := lc.Plugins["gsm"]
	if !ok {
		t.Fatalf("expected a \"gsm\" entry in Plugins")
	}
	if gsm.Url != "https://example.com/lookup_gsm.wasm" {
		t.Errorf("gsm.Url = %q, want the configured URL", gsm.Url)
	}
	if gsm.PluginConfig["credentials"] != "env:GOOGLE_APPLICATION_CREDENTIALS" {
		t.Errorf("gsm.PluginConfig[\"credentials\"] = %q, want env:GOOGLE_APPLICATION_CREDENTIALS", gsm.PluginConfig["credentials"])
	}
	if _, ok := lc.Plugins["vault"]; !ok {
		t.Errorf("expected a \"vault\" entry in Plugins")
	}
}

// TestLookupConfig_ViperUnmarshal_SeparatesDisabledFromPlugins verifies the
// same "disabled" field / remaining-provider-map split under viper's
// mapstructure-based decoding, which is the path the main gdg.yml config
// goes through (readViperConfig -> viper.Unmarshal) — a different decode
// engine than the yaml.v3 path above, and the "mapstructure:,remain" tag
// needs to behave the same way "yaml:,inline" does.
func TestLookupConfig_ViperUnmarshal_SeparatesDisabledFromPlugins(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	raw := []byte(`
disabled: true
gsm:
  url: https://example.com/lookup_gsm.wasm
  config:
    credentials: env:GOOGLE_APPLICATION_CREDENTIALS
`)
	if err := v.ReadConfig(bytes.NewReader(raw)); err != nil {
		t.Fatalf("viper ReadConfig failed: %v", err)
	}

	var lc LookupConfig
	if err := v.Unmarshal(&lc); err != nil {
		t.Fatalf("viper Unmarshal failed: %v", err)
	}

	if !lc.Disabled {
		t.Errorf("expected Disabled=true, got false")
	}
	if _, ok := lc.Plugins["disabled"]; ok {
		t.Errorf("\"disabled\" leaked into Plugins map under mapstructure decoding")
	}
	gsm, ok := lc.Plugins["gsm"]
	if !ok {
		t.Fatalf("expected a \"gsm\" entry in Plugins")
	}
	if gsm.Url != "https://example.com/lookup_gsm.wasm" {
		t.Errorf("gsm.Url = %q, want the configured URL", gsm.Url)
	}
}
