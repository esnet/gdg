package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newTestEntry() PluginRegistryEntry {
	return PluginRegistryEntry{
		Name:        "aes-256-gcm",
		Type:        PluginTypeCipher,
		Description: "seeded implementation of aes-256",
		Source:      "https://github.com/esnet/gdg-plugins/tree/main/cipher/aes-256-gcm",
		URLPattern:  "https://github.com/esnet/gdg-plugins/raw/refs/tags/{version}/plugins/cipher_aes256_gcm.wasm",
		Versions: []PluginVersionEntry{
			{Version: "0.1.0", ConfigFields: []string{"passphrase"}},
			{Version: "0.2.0", ConfigFields: []string{"passphrase", "iterations"}},
		},
	}
}

func TestResolveURL(t *testing.T) {
	entry := newTestEntry()
	got := entry.ResolveURL("0.1.0")
	assert.Equal(t,
		"https://github.com/esnet/gdg-plugins/raw/refs/tags/0.1.0/plugins/cipher_aes256_gcm.wasm",
		got,
	)
}

func TestResolveURL_MultipleVersionTokens(t *testing.T) {
	entry := PluginRegistryEntry{URLPattern: "https://example.com/{version}/plugin-{version}.wasm"}
	got := entry.ResolveURL("1.2.3")
	assert.Equal(t, "https://example.com/1.2.3/plugin-1.2.3.wasm", got)
}

func TestLatestVersion(t *testing.T) {
	entry := newTestEntry()
	latest := entry.LatestVersion()
	assert.NotNil(t, latest)
	assert.Equal(t, "0.2.0", latest.Version)
}

func TestLatestVersion_Empty(t *testing.T) {
	entry := PluginRegistryEntry{}
	assert.Nil(t, entry.LatestVersion())
}

func TestFindVersion_Found(t *testing.T) {
	entry := newTestEntry()
	v := entry.FindVersion("0.1.0")
	assert.NotNil(t, v)
	assert.Equal(t, []string{"passphrase"}, v.ConfigFields)
}

func TestFindVersion_NotFound(t *testing.T) {
	entry := newTestEntry()
	assert.Nil(t, entry.FindVersion("9.9.9"))
}

func TestPluginTypeCipherConstant(t *testing.T) {
	assert.Equal(t, "cipher", PluginTypeCipher)
}
