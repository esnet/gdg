package domain

import "strings"

const (
	// PluginTypeCipher is the only plugin type currently supported.
	PluginTypeCipher = "cipher"

	// RegistryDefaultURL is the canonical location of the GDG plugin registry JSON file.
	RegistryDefaultURL = "https://raw.githubusercontent.com/esnet/gdg-plugins/refs/heads/main/plugin_registry.json"
)

// PluginRegistryEntry represents a single plugin entry in the remote registry JSON array.
type PluginRegistryEntry struct {
	Name        string               `json:"name"`
	Type        string               `json:"type"`
	Description string               `json:"description"`
	Source      string               `json:"source"`
	URLPattern  string               `json:"urlPattern"`
	Versions    []PluginVersionEntry `json:"versions"`
}

// PluginVersionEntry describes one released version of a plugin and which
// config fields the caller must supply when using it.
type PluginVersionEntry struct {
	Version      string   `json:"version"`
	ConfigFields []string `json:"config_fields"`
}

// ResolveURL returns the concrete WASM download URL for this entry by replacing
// the "{version}" placeholder in URLPattern with the given version string.
func (e *PluginRegistryEntry) ResolveURL(version string) string {
	return strings.ReplaceAll(e.URLPattern, "{version}", version)
}

// LatestVersion returns the last element of Versions, which is assumed to be
// the most recently released version. Returns nil if Versions is empty.
func (e *PluginRegistryEntry) LatestVersion() *PluginVersionEntry {
	if len(e.Versions) == 0 {
		return nil
	}
	return &e.Versions[len(e.Versions)-1]
}

// FindVersion returns the PluginVersionEntry whose Version field matches the
// given string, or nil if no matching entry is found.
func (e *PluginRegistryEntry) FindVersion(version string) *PluginVersionEntry {
	for i := range e.Versions {
		if e.Versions[i].Version == version {
			return &e.Versions[i]
		}
	}
	return nil
}
