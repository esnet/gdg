package secure

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/esnet/gdg/internal/config/domain"
	resourceTypes "github.com/esnet/gdg/pkg/config/domain"
	"github.com/esnet/gdg/pkg/plugins/secure/contract"
	extism "github.com/extism/go-sdk"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type PluginCipherEncoder struct {
	cfg          *domain.PluginEntity
	secureFields map[string][]string
	wasmExec     *extism.Plugin
}

// EncodeValue encodes the input byte slice using a WebAssembly plugin and returns the encoded bytes or an error.
func (p PluginCipherEncoder) EncodeValue(b string) (string, error) {
	exit, out, err := p.wasmExec.Call(contract.EncodeOperation, []byte(b))
	if err != nil {
		return "", err
	}
	if exit != 0 {
		return "", fmt.Errorf("plugin returned non-zero exit code, failed to encode")
	}

	return string(out), nil
}

// DecodeValue decodes the input byte slice using a WebAssembly plugin and returns decoded bytes or an error.
func (p PluginCipherEncoder) DecodeValue(b string) (string, error) {
	exit, out, err := p.wasmExec.Call(contract.DecodeOperation, []byte(b))
	if err != nil {
		return "", err
	}
	if exit != 0 {
		return "", fmt.Errorf("plugin returned non-zero exit code, failed to encode")
	}

	return string(out), nil
}

// updateJson applies the provided fn to each secure field of the JSON byte slice b for the given resourceType.
// It returns a new JSON byte slice with transformed values or the original if no secure fields exist.
// Errors during fn execution are logged and skipped; errors from sjson.SetBytes are also logged.
func (p PluginCipherEncoder) updateJson(resourceType resourceTypes.ResourceType, b []byte, fn func(string) (string, error)) []byte {
	if p.secureFields[resourceType.String()] == nil {
		return b
	}

	for _, field := range p.secureFields[resourceType.String()] {
		result := gjson.GetBytes(b, field)
		if !result.Exists() {
			continue
		}

		// Collect all paths that need updating
		var updates []struct {
			path  string
			value string
		}

		// Track indices for each # in the pattern
		var indices []int
		collectPaths(field, result, indices, &updates, fn)

		slog.Debug("Updates collected", "count", len(updates))

		// Apply all updates
		for _, update := range updates {
			newJson, err := sjson.SetBytes(b, update.path, update.value)
			if err != nil {
				slog.Error("Failed to set value", "path", update.path, "err", err)
				continue
			}
			b = newJson
		}
	}

	return b
}

// collectPaths traverses a JSON value matching pattern, building concrete paths by replacing '#' with indices and applying fn to leaf values, appending results to updates slice.
func collectPaths(pattern string, result gjson.Result, indices []int, updates *[]struct{ path, value string }, fn func(string) (string, error)) {
	slog.Debug("collectPaths", "pattern", pattern, "isArray", result.IsArray(), "type", result.Type, "indicesLen", len(indices), "value", result.Raw)

	// If result is an empty array, nothing to process
	if result.IsArray() && len(result.Array()) == 0 {
		slog.Debug("Empty array, skipping")
		return
	}

	if !result.IsArray() {
		// Leaf node - build the final path
		finalPath := pattern
		for _, idx := range indices {
			finalPath = strings.Replace(finalPath, "#", fmt.Sprintf("%d", idx), 1)
		}

		slog.Debug("Leaf found", "path", finalPath, "value", result.Raw)

		newVal, err := fn(result.String())
		if err != nil {
			slog.Error("Failed to encode/decode value", "val", result.Raw, "err", err)
			return
		}
		*updates = append(*updates, struct{ path, value string }{finalPath, newVal})
		return
	}

	// Check if this is an array but we've already collected enough indices
	// This means we should treat the array elements as leaf values
	numWildcards := strings.Count(pattern, "#")
	if len(indices) >= numWildcards {
		// We have enough indices - the result should be leaf values
		slog.Debug("Have enough indices, treating as leaf", "numWildcards", numWildcards, "indicesCount", len(indices))

		// Build the final path
		finalPath := pattern
		for _, idx := range indices {
			finalPath = strings.Replace(finalPath, "#", fmt.Sprintf("%d", idx), 1)
		}

		slog.Debug("Leaf found (array case)", "path", finalPath, "value", result.Raw)

		newVal, err := fn(result.String())
		if err != nil {
			slog.Error("Failed to encode value", "val", result.Raw, "err", err)
			return
		}
		*updates = append(*updates, struct{ path, value string }{finalPath, newVal})
		return
	}

	// Iterate array and recurse
	for idx, item := range result.Array() {
		newIndices := append(append([]int{}, indices...), idx) // Make a copy of indices
		collectPaths(pattern, item, newIndices, updates, fn)
	}
}

// Encode applies the encoder to secure fields in JSON for the given resource type.
func (p PluginCipherEncoder) Encode(resourceType resourceTypes.ResourceType, b []byte) ([]byte, error) {
	b = p.updateJson(resourceType, b, p.EncodeValue)
	return b, nil
}

// Decode applies the decoder to secure fields in JSON for the given resource type.
func (p PluginCipherEncoder) Decode(resourceType resourceTypes.ResourceType, b []byte) ([]byte, error) {
	b = p.updateJson(resourceType, b, p.DecodeValue)
	return b, nil
}

// NewPluginCipherEncoder creates a CipherEncoder that uses a WebAssembly plugin to encode and decode values.
// It initializes the plugin from either a file path or URL, processes configuration,
// and returns a PluginCipherEncoder ready for use.
func NewPluginCipherEncoder(plugCfg *domain.PluginEntity, secureFields map[string][]string) contract.CipherEncoder {
	o := &PluginCipherEncoder{
		cfg:          plugCfg,
		secureFields: secureFields,
	}
	var wasmInt extism.Wasm
	if plugCfg.FilePath != "" {
		wasmInt = extism.WasmFile{Path: plugCfg.FilePath}
	} else if plugCfg.Url != "" {
		wasmInt = extism.WasmUrl{Url: plugCfg.Url}
	} else {
		log.Fatal("plugin configuration is invalid. No Url or file path was found")
	}

	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			wasmInt,
		},
		Config: plugCfg.GetPluginConfig(),
	}

	ctx := context.Background()
	config := extism.PluginConfig{
		EnableWasi: true,
	}
	plugin, err := extism.NewPlugin(ctx, manifest, config, []extism.HostFunction{})
	if err != nil {
		fmt.Printf("Failed to initialize plugin: %v\n", err)
		os.Exit(1)
	}

	o.wasmExec = plugin
	return o
}
