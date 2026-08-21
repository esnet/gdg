package domain

import (
	"testing"

	"github.com/esnet/gdg/pkg/version"
	"github.com/stretchr/testify/assert"
)

// withGdgVersion temporarily overrides the package-level version.Version
// (normally set by the linker at release build time, and "DEVEL" otherwise)
// so IsValid's range-check logic can be exercised deterministically. The
// original value is restored after the test.
func withGdgVersion(t *testing.T, v string) {
	t.Helper()
	original := version.Version
	version.Version = v
	t.Cleanup(func() {
		version.Version = original
	})
}

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

// ── IsValid ───────────────────────────────────────────────────────────────────

func TestIsValid_DevelBypassesRangeCheck(t *testing.T) {
	withGdgVersion(t, "DEVEL")
	// A range that would reject any real version; DEVEL should short-circuit
	// past it and report valid regardless.
	e := PluginVersionEntry{MinimumVersion: "99.0.0", MaximumVersion: "99.0.0"}
	assert.True(t, e.IsValid())
}

func TestIsValid_WithinRange(t *testing.T) {
	withGdgVersion(t, "1.5.0")
	e := PluginVersionEntry{MinimumVersion: "1.0.0", MaximumVersion: "2.0.0"}
	assert.True(t, e.IsValid())
}

func TestIsValid_AtRangeBoundsInclusive(t *testing.T) {
	withGdgVersion(t, "1.0.0")
	min := PluginVersionEntry{MinimumVersion: "1.0.0", MaximumVersion: "2.0.0"}
	assert.True(t, min.IsValid(), "minimum bound should be inclusive")

	withGdgVersion(t, "2.0.0")
	max := PluginVersionEntry{MinimumVersion: "1.0.0", MaximumVersion: "2.0.0"}
	assert.True(t, max.IsValid(), "maximum bound should be inclusive")
}

func TestIsValid_BelowMinimum(t *testing.T) {
	withGdgVersion(t, "0.9.0")
	e := PluginVersionEntry{MinimumVersion: "1.0.0", MaximumVersion: "2.0.0"}
	assert.False(t, e.IsValid())
}

func TestIsValid_AboveMaximum(t *testing.T) {
	withGdgVersion(t, "3.0.0")
	e := PluginVersionEntry{MinimumVersion: "1.0.0", MaximumVersion: "2.0.0"}
	assert.False(t, e.IsValid())
}

func TestIsValid_NoBoundsIsAlwaysValid(t *testing.T) {
	withGdgVersion(t, "0.0.1")
	e := PluginVersionEntry{}
	assert.True(t, e.IsValid())
}

func TestIsValid_MalformedBoundIsInvalid(t *testing.T) {
	withGdgVersion(t, "1.0.0")
	e := PluginVersionEntry{MinimumVersion: "not-a-semver-version"}
	assert.False(t, e.IsValid())
}
