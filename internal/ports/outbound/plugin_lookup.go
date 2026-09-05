package outbound

import "strings"

// LookupOperation is the exported WASM guest function name a lookup plugin
// must implement, invoked via extism.Plugin.Call.
const LookupOperation = "Lookup"

// LookupPrefix marks a config value (e.g. a Token or Password field) as a
// dynamic lookup reference — "lookup:<name>:<key>[.<json_field>]" — rather
// than a literal value or cipher-encoded ciphertext.
//
// Defined here (rather than in the lookup adapter package itself) so that
// both config_domain (which needs it to decide whether a secure field
// should skip cipher decoding) and the lookup adapter package (which
// parses and resolves it) can depend on the same constant without a
// circular import: the lookup package already imports config_domain for
// PluginEntity/PluginConfig, so config_domain cannot import it back.
const LookupPrefix = "lookup:"

// IsLookupRef reports whether raw is a lookup reference (i.e. starts with
// LookupPrefix).
func IsLookupRef(raw string) bool {
	return strings.HasPrefix(raw, LookupPrefix)
}

// LookupProvider resolves a single key (e.g. a Vault path or a GCP Secret
// Manager resource name) into its concrete secret value. Implementations
// may be backed by a WASM plugin (via extism) or, in principle, any other
// mechanism that satisfies this interface.
type LookupProvider interface {
	// Lookup resolves key to its secret value, or returns an error if the
	// key cannot be resolved (not found, auth failure, network error, etc.).
	Lookup(key string) (string, error)
}

// LookupResolver resolves a full "lookup:<name>:<key>[.<json_field>]"
// reference (see the lookup adapter package) to its concrete value,
// dispatching to the correct named LookupProvider and applying any
// ".json_field" extraction. The concrete implementation is
// lookup.Resolver; this interface exists so higher layers (e.g.
// GrafanaConfig, DashNGoImpl) can depend on the behavior without
// importing the adapter package directly.
type LookupResolver interface {
	Resolve(raw string) (string, error)
}
