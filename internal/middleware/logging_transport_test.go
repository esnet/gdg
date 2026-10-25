package middleware

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type handler struct{}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello test"))
}

type errExecuteHTTPRequestTransport struct{}

func (e *errExecuteHTTPRequestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, errors.New("an error during HTTP request execution")
}

type errNilResponseTransport struct{}

func (e *errNilResponseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, nil
}
func TestLoggingTransportWrapper(t *testing.T) {
	server := httptest.NewTestServer(t, handler{})
	client := server.Client()

	testCases := []struct {
		name           string
		includeHeaders bool
		includeBody    bool
		inner          http.RoundTripper
		request        *http.Request
		expectedErr    error
	}{
		{
			name:           "success",
			includeHeaders: true,
			includeBody:    true,
			inner:          client.Transport,
			request:        httptest.NewRequest("POST", server.URL, strings.NewReader("test body")),
			expectedErr:    nil,
		},
		{
			name:           "disabled_logging_success",
			includeHeaders: false,
			includeBody:    false,
			inner:          client.Transport,
			request:        httptest.NewRequest("POST", server.URL, strings.NewReader("test body")),
			expectedErr:    nil,
		},
		{
			name:           "nil_request",
			includeHeaders: true,
			includeBody:    false,
			inner:          client.Transport,
			request:        nil,
			expectedErr:    ErrNilRequest,
		},
		{
			name:           "invalid_request",
			includeHeaders: false,
			includeBody:    true,
			inner:          client.Transport,
			request: &http.Request{
				URL:    nil,
				Header: nil,
			},
			expectedErr: ErrInvalidRequest,
		},
		{
			name:           "http_request_execution_fails",
			includeHeaders: true,
			includeBody:    true,
			inner:          &errExecuteHTTPRequestTransport{},
			request:        httptest.NewRequest("POST", server.URL, strings.NewReader("test body")),
			expectedErr:    ErrExecuteHTTPRequest,
		},
		{
			name:           "nil_response_without_inner_error",
			includeHeaders: true,
			includeBody:    false,
			inner:          &errNilResponseTransport{},
			request:        httptest.NewRequest("POST", server.URL, strings.NewReader("test body")),
			expectedErr:    ErrNilResponse,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := log.New(&buf, "", 0)
			wrapper := NewLoggingTransportWrapper(logger, tc.includeHeaders, tc.includeBody)

			rt := wrapper(tc.inner)

			resp, err := rt.RoundTrip(tc.request)
			if err != nil {
				assert.Nil(t, resp)
			} else {
				assert.NotNil(t, resp)
				// if logging disabled, no middleware logs must be produced
				if !tc.includeHeaders && !tc.includeBody {
					assert.Zero(t, buf.Len())
				} else {
					assert.NotZero(t, buf.Len())
				}
			}

			assert.ErrorIs(t, err, tc.expectedErr)

			if tc.expectedErr == nil {
				assert.Equal(t, http.StatusOK, resp.StatusCode)
			}
		})
	}
}

func TestLoggingTransportWrapper_nilInputs(t *testing.T) {
	t.Run("logging_enabled", func(t *testing.T) {
		wrapper := NewLoggingTransportWrapper(nil, true, true)
		transport := wrapper(nil)

		logger := log.Default()
		original := logger.Writer()

		logger.SetOutput(io.Discard)
		t.Cleanup(func() {
			logger.SetOutput(original)
		})

		assert.NotPanics(t, func() {
			transport.RoundTrip(httptest.NewRequest("POST", "http://www.example.org", strings.NewReader("test body")))
		})
	})

	t.Run("logging_disabled", func(t *testing.T) {
		wrapper := NewLoggingTransportWrapper(nil, false, false)
		transport := wrapper(nil)

		if tr, ok := transport.(*LoggingTransport); ok {
			assert.Equal(t, tr.logger, log.Default())
		}

		assert.NotPanics(t, func() {
			transport.RoundTrip(httptest.NewRequest("POST", "http://www.example.org", strings.NewReader("test body")))
		})
	})
}

func TestFormatDump(t *testing.T) {
	const body = "Go is a general-purpose language designed with systems programming in mind."
	req, err := http.NewRequest("PUT", "http://www.example.org", strings.NewReader(body))
	if err != nil {
		log.Fatal(err)
	}

	validDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		log.Fatal(err)
	}

	testCases := []struct {
		name        string
		dump        []byte
		expectedErr error
	}{
		{
			name:        "success",
			dump:        validDump,
			expectedErr: nil,
		},
		{
			name:        "empty_dump",
			dump:        []byte{},
			expectedErr: ErrInvalidDump,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := formatDump(tc.dump, true, true)
			assert.ErrorIs(t, err, tc.expectedErr)
		})
	}
}
