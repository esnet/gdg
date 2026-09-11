package gsm

import (
	"errors"
	"testing"
	"time"

	extism "github.com/extism/go-sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// fakeTokenSource is a minimal oauth2.TokenSource for testing, returning
// either a fixed token or a fixed error.
type fakeTokenSource struct {
	token *oauth2.Token
	err   error
}

func (f fakeTokenSource) Token() (*oauth2.Token, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.token, nil
}

// ── fetchAccessToken ────────────────────────────────────────────────────

func TestFetchAccessToken_Success(t *testing.T) {
	ts := fakeTokenSource{token: &oauth2.Token{
		AccessToken: "test-access-token",
		Expiry:      time.Now().Add(time.Hour),
	}}

	tok, err := fetchAccessToken(ts)
	require.NoError(t, err)
	assert.Equal(t, "test-access-token", tok)
}

func TestFetchAccessToken_PropagatesTokenSourceError(t *testing.T) {
	ts := fakeTokenSource{err: errors.New("token exchange failed")}

	_, err := fetchAccessToken(ts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to obtain GCP access token")
	assert.Contains(t, err.Error(), "token exchange failed")
}

// fakeMemoryWriter is a minimal tokenMemoryWriter for testing
// handleGetAccessToken without a real WASM guest.
type fakeMemoryWriter struct {
	offset uint64
	err    error
}

func (f fakeMemoryWriter) WriteString(_ string) (uint64, error) {
	if f.err != nil {
		return 0, f.err
	}
	return f.offset, nil
}

// ── handleGetAccessToken ────────────────────────────────────────────────

func TestHandleGetAccessToken_Success_ReturnsWrittenOffset(t *testing.T) {
	ts := fakeTokenSource{token: &oauth2.Token{AccessToken: "test-access-token"}}
	w := fakeMemoryWriter{offset: 42}

	got := handleGetAccessToken(ts, w)
	assert.Equal(t, uint64(42), got)
}

func TestHandleGetAccessToken_TokenSourceError_ReturnsZero(t *testing.T) {
	ts := fakeTokenSource{err: errors.New("token exchange failed")}
	w := fakeMemoryWriter{offset: 99} // should never be reached

	got := handleGetAccessToken(ts, w)
	assert.Equal(t, uint64(0), got, "a token-source error must yield offset 0, not the writer's offset")
}

func TestHandleGetAccessToken_WriteError_ReturnsZero(t *testing.T) {
	ts := fakeTokenSource{token: &oauth2.Token{AccessToken: "test-access-token"}}
	w := fakeMemoryWriter{err: errors.New("guest memory full")}

	got := handleGetAccessToken(ts, w)
	assert.Equal(t, uint64(0), got)
}

// ── newGetAccessTokenHostFunction ───────────────────────────────────────
//
// The real stack-callback closure (the one-line adapter wrapping
// handleGetAccessToken with a live *extism.CurrentPlugin) needs a running
// WASM guest that actually imports and calls this host function to
// exercise end-to-end — that only exists once the real gdg-plugins guest
// plugin is built (see gsm_plan.md step 11). What's verified here instead
// is that the constructed HostFunction carries the ABI a guest will
// expect: correct exported name, default extism host-function namespace,
// no parameters, and a single pointer-typed return (the offset of the
// written token string). Combined with the handleGetAccessToken tests
// above, this covers all of the function's actual logic — the only thing
// left untested is the single-line stack[0]-assignment glue itself.

func TestNewGetAccessTokenHostFunction_HasExpectedSignature(t *testing.T) {
	hostFn := newGetAccessTokenHostFunction(fakeTokenSource{})

	assert.Equal(t, "get_gcp_access_token", hostFn.Name)
	assert.Equal(t, "extism:host/user", hostFn.Namespace)
	assert.Empty(t, hostFn.Params, "host function should take no parameters")
	require.Len(t, hostFn.Returns, 1, "host function should return exactly one value (the token string's memory offset)")
	assert.Equal(t, extism.ValueTypePTR, hostFn.Returns[0], "return value should be a memory pointer/offset")
}
