package domain

import (
	"net/http"
)

// TransportWrapper wraps RoundTripper with middleware code.
type TransportWrapper func(rt http.RoundTripper) http.RoundTripper
