// Package registry provides a client for loading and querying the GDG plugin registry.
// The registry is a JSON array of PluginRegistryEntry values hosted at a remote URL or
// available as a local file. The client fetches the data once per process invocation and
// caches the result in memory for all subsequent calls.
package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/esnet/gdg/internal/domain"
)

// ClientConfig controls how the registry is loaded.
// FilePath takes precedence over URL when both are set.
type ClientConfig struct {
	// FilePath, if non-empty, reads the registry JSON from disk instead of
	// fetching it over the network. Useful for air-gapped environments or
	// local development of new plugins.
	FilePath string

	// URL is the remote registry endpoint. Defaults to domain.RegistryDefaultURL
	// when empty.
	URL string

	// HTTPClient allows injection of a custom HTTP client (e.g., httptest in tests).
	// When nil, http.DefaultClient is used.
	HTTPClient *http.Client
}

// Client loads and caches the plugin registry for the lifetime of one process invocation.
type Client struct {
	cfg      ClientConfig
	once     sync.Once
	cached   []domain.PluginRegistryEntry
	fetchErr error
}

// NewClient returns a new Client configured with cfg.
func NewClient(cfg ClientConfig) *Client {
	return &Client{cfg: cfg}
}

// All returns all registry entries, loading and caching on the first call.
// Subsequent calls return the cached slice without any I/O.
func (c *Client) All() ([]domain.PluginRegistryEntry, error) {
	c.once.Do(func() {
		c.cached, c.fetchErr = c.load()
	})
	return c.cached, c.fetchErr
}

// CipherPlugins returns only entries whose Type equals domain.PluginTypeCipher.
func (c *Client) CipherPlugins() ([]domain.PluginRegistryEntry, error) {
	all, err := c.All()
	if err != nil {
		return nil, err
	}
	var result []domain.PluginRegistryEntry
	for _, e := range all {
		if e.Type == domain.PluginTypeCipher {
			result = append(result, e)
		}
	}
	return result, nil
}

// Find returns the registry entry whose Name matches name (case-insensitive),
// or an error if no such entry exists.
func (c *Client) Find(name string) (*domain.PluginRegistryEntry, error) {
	all, err := c.All()
	if err != nil {
		return nil, err
	}
	lower := strings.ToLower(name)
	for i := range all {
		if strings.ToLower(all[i].Name) == lower {
			return &all[i], nil
		}
	}
	return nil, fmt.Errorf("plugin %q not found in registry", name)
}

// ResolvePlugin looks up name and version in the registry and returns the matched
// entry, its version metadata, and the resolved WASM URL. If version is empty,
// the latest version is used. Returns an error if the plugin or version is not found.
func (c *Client) ResolvePlugin(name, version string) (*domain.PluginRegistryEntry, *domain.PluginVersionEntry, string, error) {
	entry, err := c.Find(name)
	if err != nil {
		return nil, nil, "", err
	}

	var versionEntry *domain.PluginVersionEntry
	if version == "" {
		versionEntry = entry.LatestVersion()
		if versionEntry == nil {
			return nil, nil, "", fmt.Errorf("plugin %q has no versions in the registry", name)
		}
	} else {
		versionEntry = entry.FindVersion(version)
		if versionEntry == nil {
			return nil, nil, "", fmt.Errorf("plugin %q version %q not found in registry", name, version)
		}
	}

	wasmURL := entry.ResolveURL(versionEntry.Version)
	return entry, versionEntry, wasmURL, nil
}

// load reads the registry from a local file (if ClientConfig.FilePath is set)
// or fetches it from the configured URL.
func (c *Client) load() ([]domain.PluginRegistryEntry, error) {
	var data []byte
	var err error

	if c.cfg.FilePath != "" {
		data, err = os.ReadFile(c.cfg.FilePath) // #nosec G304
		if err != nil {
			return nil, fmt.Errorf("reading registry file %q: %w", c.cfg.FilePath, err)
		}
	} else {
		url := c.cfg.URL
		if url == "" {
			url = domain.RegistryDefaultURL
		}
		data, err = c.fetch(url)
		if err != nil {
			return nil, err
		}
	}

	var entries []domain.PluginRegistryEntry
	if err = json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parsing registry JSON: %w", err)
	}
	return entries, nil
}

// fetch performs an HTTP GET against url using the configured HTTPClient
// and returns the response body.
func (c *Client) fetch(url string) ([]byte, error) {
	httpClient := c.cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	resp, err := httpClient.Get(url) // #nosec G107
	if err != nil {
		return nil, fmt.Errorf("fetching registry from %q: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching registry from %q: HTTP %d", url, resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading registry response body: %w", err)
	}
	return body, nil
}
