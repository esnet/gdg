package middleware

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/esnet/gdg/internal/domain"
)

var (
	ErrNilRequest         = errors.New("nil HTTP request")
	ErrInvalidRequest     = errors.New("invalid request")
	ErrExecuteHTTPRequest = errors.New("execute HTTP request")
	ErrNilResponse        = errors.New("nil HTTP response")

	ErrInvalidDump = errors.New("invalid dump")
)

type LoggingTransport struct {
	logger         *log.Logger
	inner          http.RoundTripper
	includeHeaders bool
	includeBody    bool
}

// NewLoggingTransportWrapper creates HTTP traffic logging
// middleware used specifically to wrap http.Transport.
func NewLoggingTransportWrapper(logger *log.Logger, includeHeaders, includeBody bool) domain.TransportWrapper {
	if !includeHeaders && !includeBody {
		return func(rt http.RoundTripper) http.RoundTripper {
			if rt == nil {
				rt = http.DefaultTransport
			}

			return rt
		}
	}

	if logger == nil {
		logger = log.Default()
	}

	return func(rt http.RoundTripper) http.RoundTripper {
		if !includeHeaders && !includeBody {
			return rt
		}

		if rt == nil {
			rt = http.DefaultTransport
		}

		return &LoggingTransport{
			logger:         logger,
			inner:          rt,
			includeHeaders: includeHeaders,
			includeBody:    includeBody,
		}
	}
}

// RoundTrip logs the request, executes HTTP request and logs the obtained response.
func (lt *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, ErrNilRequest
	}

	if req.URL == nil || req.Header == nil {
		return nil, ErrInvalidRequest
	}

	reqDump, err := httputil.DumpRequestOut(req, lt.includeBody)
	if err != nil {
		lt.logger.Printf("failed to dump request out: %v", err)
	} else {
		reqLog, err := formatDump(reqDump, lt.includeHeaders, lt.includeBody)
		if err != nil {
			lt.logger.Printf("failed to format request dump: %v", err)
		} else {
			lt.logger.Println(reqLog)
		}
	}

	resp, err := lt.inner.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrExecuteHTTPRequest, err)
	}

	if resp == nil {
		return nil, ErrNilResponse
	}

	respDump, err := httputil.DumpResponse(resp, lt.includeBody)
	if err != nil {
		lt.logger.Printf("failed to dump response: %v", err)
	} else {
		respLog, err := formatDump(respDump, lt.includeHeaders, lt.includeBody)
		if err != nil {
			lt.logger.Printf("failed to format response dump: %v", err)
		} else {
			lt.logger.Println(respLog)
		}

	}

	return resp, nil
}

func formatDump(dump []byte, includeHeaders, includeBody bool) (string, error) {
	if len(dump) == 0 {
		return "", ErrInvalidDump
	}

	sep := []byte("\r\n\r\n")
	before, after, found := bytes.Cut(dump, sep)
	if !found {
		return "", ErrInvalidDump
	}

	sb := strings.Builder{}

	if includeHeaders {
		sb.Write(before)
	}

	if includeHeaders && includeBody {
		sb.Write([]byte("\r\n"))
	}

	if includeBody {
		sb.Write([]byte("Body: "))
		sb.Write(after)
	}

	sb.Write([]byte("\r\n"))

	return sb.String(), nil
}
