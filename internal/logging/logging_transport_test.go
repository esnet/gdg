package logging_test

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/esnet/gdg/internal/logging"
)

type handler struct{}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello test"))
}

func TestLoggingTransport_RoundTrip_Success(t *testing.T) {
	h := handler{}
	server := httptest.NewServer(h)
	defer server.Close()

	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(io.Discard)

	client := &http.Client{
		Transport: logging.NewTransport(http.DefaultTransport, true),
	}

	req, err := http.NewRequest(http.MethodGet, server.URL, strings.NewReader("hello client"))
	require.NoError(t, err)

	res, err := client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, "hello test", string(body))

	logs := logBuf.String()
	assert.Contains(t, logs, "GET / HTTP/1.1")
	assert.Contains(t, logs, "hello client")
	assert.Contains(t, logs, "HTTP/1.1 200 OK")
	assert.Contains(t, logs, "hello test")
}

func TestLoggingTransport_NilRequest(t *testing.T) {
	tr := logging.NewTransport(nil, false)
	
	res, err := tr.RoundTrip(nil)
	assert.ErrorIs(t, err, logging.ErrNilRequest)
	assert.Nil(t, res)
}

func TestLoggingTransport_DefaultTransportFallback(t *testing.T) {
	h := handler{}
	server := httptest.NewServer(h)
	defer server.Close()

	tr := logging.NewTransport(nil, false)
	client := &http.Client{Transport: tr}

	res, err := client.Get(server.URL)
	require.NoError(t, err)
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}