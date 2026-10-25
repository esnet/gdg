package middleware

import (
	"errors"
	"net/http"
)

var ErrNilRoundTripper = errors.New("nil RoundTripper")

// TransportWrapper wraps RoundTripper with middleware code.
type TransportWrapper func(rt http.RoundTripper) http.RoundTripper
