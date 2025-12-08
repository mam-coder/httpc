// Package httpc provides HTTP client functionality.
// This file contains the DebugTransport implementation for detailed request/response logging.
package httpc

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// DebugTransport is a wrapper around http.DefaultTransport that logs HTTP requests and responses.
type DebugTransport struct {
	transport   http.RoundTripper
	Debug       bool
	Logger      *log.Logger
	LogBody     bool
	MaxBodySize int64
}

// NewDebugTransport creates a new DebugTransport with the given options.
// It wraps the provided transport in debug logging functionality.
// The default logger writes to stdout with a timestamp and microsecond precision.

func NewDebugTransport(transport http.RoundTripper, debug bool) *DebugTransport {
	if transport == nil {
		transport = defaultTransport()
	}
	return &DebugTransport{
		transport:   transport,
		Debug:       debug,
		Logger:      log.New(os.Stdout, "[DEBUG] ", log.LstdFlags|log.Lmicroseconds),
		LogBody:     true,
		MaxBodySize: 1024 * 1024,
	}
}

// RoundTrip implements the http.RoundTripper interface for DebugTransport.
// It logs the request and response and delegates to the underlying RoundTripper.
func (t *DebugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if !t.Debug {
		return t.transport.RoundTrip(req)
	}
	//Log the request
	t.Logger.Printf("→ %s %s", req.Method, req.URL)
	t.logHeaders("Request Headers", req.Header)

	if t.LogBody && req.Body != nil {
		body, err := io.ReadAll(io.LimitReader(req.Body, t.MaxBodySize))
		if err == nil {
			req.Body = io.NopCloser(bytes.NewBuffer(body))
			decodedBody := t.decodeBody(body, req.Header.Get("Content-Encoding"))
			t.Logger.Printf("Request Body:\n%s", string(decodedBody))
		}
	}

	//Execute the request
	start := time.Now()
	resp, err := t.transport.RoundTrip(req)
	duration := time.Since(start)
	if err != nil {
		t.Logger.Printf("✗ Error: %v (took %v)", err, duration)
		return resp, err
	}

	//Log the response
	t.Logger.Printf("← %d %s (took %v)", resp.StatusCode, http.StatusText(resp.StatusCode), duration)
	t.logHeaders("Response Headers", resp.Header)

	if t.LogBody && resp.Body != nil {
		body, err := io.ReadAll(io.LimitReader(resp.Body, t.MaxBodySize))
		if err == nil {
			resp.Body = io.NopCloser(bytes.NewBuffer(body))
			decodedBody := t.decodeBody(body, resp.Header.Get("Content-Encoding"))
			t.Logger.Printf("Response Body:\n%s", string(decodedBody))
		}
	}

	return resp, nil
}

// logHeaders logs the headers in the given title.
// It filters out sensitive headers.
// It logs each header key and value on a separate line.
// It ignores empty headers.

func (t *DebugTransport) logHeaders(title string, headers http.Header) {
	if len(headers) == 0 {
		return
	}

	t.Logger.Printf("%s:", title)

	for key, values := range headers {
		for _, value := range values {
			if t.isSensitive(key) {
				value = "***MODIFIED***SENSITIVE HEADER***"
			}
			t.Logger.Printf("	%s: %s", key, value)
		}
	}
}

// isSensitive returns true if the header is considered sensitive and should not be logged.
// It checks against a list of common sensitive headers.
// It is case-insensitive.
func (t *DebugTransport) isSensitive(header string) bool {
	sensitiveHeaders := []string{"authorization", "api-key", "x-api-key", "cookie"}
	lowerCaseHeader := strings.ToLower(header)
	for _, sensitiveHeader := range sensitiveHeaders {
		if strings.Contains(lowerCaseHeader, sensitiveHeader) {
			return true
		}
	}

	return false
}

// decodeBody decodes the body based on the Content-Encoding header.
// It supports gzip encoding and returns the decoded body.
// If decoding fails or encoding is not supported, it returns the original body.
func (t *DebugTransport) decodeBody(body []byte, contentEncoding string) []byte {
	if contentEncoding == "" || len(body) == 0 {
		return body
	}

	if isGzipEncoded(contentEncoding) {
		decodedBody, err := decodeGzipBody(body)
		if err != nil {
			t.Logger.Printf("Failed to decode gzip body: %v", err)
			return body
		}
		return decodedBody
	}

	return body
}
