package outbound

// LookupService is the port through which higher layers (config_domain,
// etc.) determine whether a raw config value is a lookup reference and
// what prefix marks it as such. The concrete implementation is
// lookup.Resolver; this interface exists so config_domain can depend on
// the behaviour without importing the adapter package.
type LookupService interface {
	// IsLookupRef reports whether raw is a lookup reference (e.g. starts
	// with "lookup:").
	IsLookupRef(raw string) bool
	// Prefix returns the prefix string that marks a value as a lookup
	// reference (e.g. "lookup:").
	Prefix() string
}

// LookupProvider resolves a single key (e.g. a Vault path or a GCP Secret
// Manager resource name) into its concrete secret value. Implementations
// may be backed by a WASM plugin (via extism) or any other mechanism that
// satisfies this interface.
type LookupProvider interface {
	// Lookup resolves key to its secret value, or returns an error if the
	// key cannot be resolved (not found, auth failure, network error, etc.).
	Lookup(key string) (string, error)
}

// LookupResolver resolves a full "lookup:<name>:<key>[.<json_field>]"
// reference to its concrete value, dispatching to the correct named
// LookupProvider and applying any ".json_field" extraction. The concrete
// implementation is lookup.Resolver; this interface exists so higher
// layers (e.g. GrafanaConfig, DashNGoImpl) can depend on the behavior
// without importing the adapter package directly.
type LookupResolver interface {
	Resolve(raw string) (string, error)
}
