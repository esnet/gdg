package logging

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

var (
	ErrNilRequest  = errors.New("nil HTTP request")
	ErrNilResponse = errors.New("nil HTTP response")
)

type LoggingTransport struct {
	inner       http.RoundTripper
	includeBody bool
}

func NewTransport(inner http.RoundTripper, includeBody bool) http.RoundTripper {
	if inner == nil {
		inner = http.DefaultTransport
	}
	return &LoggingTransport{
		inner:       inner,
		includeBody: includeBody,
	}
}

func (lt *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	dump, err := httputil.DumpRequestOut(req, lt.includeBody)
	if err != nil {
		return nil, fmt.Errorf("dumping request: %w", err)
	}

	reqLog := formatDump(dump)

	log.Printf("%v\n", reqLog)

	resp, err := lt.inner.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("obtaining response: %w", err)
	}

	dump, err = httputil.DumpResponse(resp, lt.includeBody)
	if err != nil {
		return nil, fmt.Errorf("dumping response: %w", err)
	}

	if resp == nil {
		return nil, ErrNilResponse
	}

	respLog := formatDump(dump)

	log.Printf("%v\n", respLog)

	return resp, nil
}

func formatDump(dump []byte) string {
	return string(bytes.Replace(dump, []byte("\n\n"), []byte("\nBody:"), 1))
}
