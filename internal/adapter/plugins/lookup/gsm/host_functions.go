package gsm

import (
	"context"
	"fmt"
	"log/slog"

	extism "github.com/extism/go-sdk"
	"golang.org/x/oauth2"
)

// getGCPAccessTokenFunctionName is the extism host function name a GSM
// lookup guest plugin calls to obtain a valid bearer token immediately
// before making its request to the Secret Manager REST API.
const getGCPAccessTokenFunctionName = "get_gcp_access_token" // #nosec G101

// fetchAccessToken retrieves the current access token string from
// tokenSource.
//
// This is split out from the extism host-function callback below on
// purpose: the callback itself needs a live *extism.CurrentPlugin backed
// by a running WASM guest to do anything (WriteString allocates and
// writes into that guest's memory), which this sandboxed environment
// can't exercise without a compiled guest plugin. Keeping the actual
// token-retrieval/error-handling logic in a plain function means it stays
// fully unit-testable against a fake oauth2.TokenSource regardless — the
// glue code around it (newGetAccessTokenHostFunction) is thin enough that
// there's little left to get wrong once this part is verified.
func fetchAccessToken(tokenSource oauth2.TokenSource) (string, error) {
	tok, err := tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("gsm lookup plugin: failed to obtain GCP access token: %w", err)
	}
	return tok.AccessToken, nil
}

// tokenMemoryWriter is the subset of *extism.CurrentPlugin's behavior the
// get_gcp_access_token host function needs: writing a string into the
// calling guest's memory and getting back its offset.
//
// Extracted as its own interface (rather than depending on
// *extism.CurrentPlugin directly) purely for testability: a real
// *extism.CurrentPlugin can only be constructed by a live, running WASM
// guest, but Go interfaces are satisfied structurally, so a fake
// implementing just WriteString lets handleGetAccessToken's logic be
// unit-tested without any WASM runtime involved.
type tokenMemoryWriter interface {
	WriteString(s string) (uint64, error)
}

// handleGetAccessToken contains the get_gcp_access_token host function's
// entire behavior: fetch a token, write it into guest memory, and return
// the offset to place on the call stack (or 0 on any failure, which the
// guest must treat as "no token available" — the failure reason itself is
// logged host-side, since there's no richer way to communicate it back
// over this simple stack-based ABI).
//
// This is deliberately where all the real logic lives. The
// extism.HostFunction callback in newGetAccessTokenHostFunction is a
// one-line adapter around it, so what actually gets unit-tested (against
// a fake tokenMemoryWriter) is this function, not something that requires
// a live WASM guest to exercise.
func handleGetAccessToken(tokenSource oauth2.TokenSource, w tokenMemoryWriter) uint64 {
	token, err := fetchAccessToken(tokenSource)
	if err != nil {
		slog.Error(err.Error())
		return 0
	}

	offset, err := w.WriteString(token)
	if err != nil {
		slog.Error("gsm lookup plugin: failed to write GCP access token to guest memory", "err", err)
		return 0
	}

	return offset
}

// newGetAccessTokenHostFunction builds the extism.HostFunction that lets
// the WASM guest ask the host for a fresh GCP access token on demand.
//
// Token acquisition and refresh stay entirely on the host: tokenSource
// (built in credentials.go from the resolved "credentials" plugin config)
// already handles expiry/refresh via golang.org/x/oauth2's own caching
// TokenSource, so the guest never has to reason about token lifetimes —
// it just calls this function right before every request and always gets
// back something usable. This is also what lets the surrounding
// extism.Plugin instance be created once and cached for the life of the
// process (the same way PluginCipherEncoder caches its wasmExec instance)
// without ever going stale: the token itself is minted fresh per call
// rather than baked into the plugin's static manifest config.
func newGetAccessTokenHostFunction(tokenSource oauth2.TokenSource) extism.HostFunction {
	return extism.NewHostFunctionWithStack(
		getGCPAccessTokenFunctionName,
		func(_ context.Context, p *extism.CurrentPlugin, stack []uint64) {
			stack[0] = handleGetAccessToken(tokenSource, p)
		},
		[]extism.ValueType{},
		[]extism.ValueType{extism.ValueTypePTR},
	)
}
