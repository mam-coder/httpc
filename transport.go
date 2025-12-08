// Package httpc provides HTTP client functionality.
// This file contains various http.RoundTripper implementations for features like
// headers, authentication, retry logic, logging, and request blocking.
package httpc

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// defaultTransport returns an optimized http.Transport with sensible defaults.
// It configures connection pooling, timeouts, HTTP/2 support, and proxy settings.
//
// Configuration details:
//   - HTTP/2 enabled with automatic fallback to HTTP/1.1
//   - 100 maximum idle connections across all hosts
//   - 10 maximum idle connections per host
//   - 30 second dial and keep-alive timeouts
//   - 90 second idle connection timeout
//   - 10 second TLS handshake timeout
//   - Compression disabled (for manual control)
//   - Proxy settings from environment variables
func defaultTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     false,
		DisableCompression:    true,
	}
}

// headerTransport is an http.RoundTripper that adds custom headers to every request.
// It wraps another RoundTripper and sets the configured headers before delegating
// to the wrapped transport.
type headerTransport struct {
	transport http.RoundTripper
	headers   map[string]string
}

// RoundTrip implements http.RoundTripper by adding headers and delegating to the wrapped transport.
func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	for key, value := range t.headers {
		req.Header.Set(key, value)
	}
	return t.transport.RoundTrip(req)
}

// baseAuthTransport is an http.RoundTripper that adds HTTP Basic Authentication
// to every request. It encodes the username and password and sets the Authorization header.
type baseAuthTransport struct {
	transport http.RoundTripper
	username  string
	password  string
}

// RoundTrip implements http.RoundTripper by adding Basic Auth and delegating to the wrapped transport.
func (t *baseAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.SetBasicAuth(t.username, t.password)

	return t.transport.RoundTrip(req)
}

// authTransport is an http.RoundTripper that adds Bearer token authentication
// to every request by setting the Authorization header with "Bearer <token>".
type authTransport struct {
	transport http.RoundTripper
	token     string
}

// RoundTrip implements http.RoundTripper by adding Bearer token auth and delegating to the wrapped transport.
func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+t.token)

	return t.transport.RoundTrip(req)
}

// blockListTransport is an http.RoundTripper that blocks requests to specific domains.
// Requests to any domain in the blockedList will fail with an error before being sent.
// This is useful for security or testing purposes.
type blockListTransport struct {
	transport   http.RoundTripper
	blockedList []string
}

// RoundTrip implements http.RoundTripper by checking the blocked list before delegating to the wrapped transport.
func (t *blockListTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	for _, blocked := range t.blockedList {
		if strings.Contains(req.URL.Host, blocked) {
			return nil, fmt.Errorf("domain %s is blocked", blocked)
		}
	}

	return t.transport.RoundTrip(req)
}

// retryTransport is an http.RoundTripper that implements automatic retry logic
// with exponential backoff. It retries failed requests based on the configured
// RetryConfig, which specifies max retries, backoff duration, and retry conditions.
type retryTransport struct {
	transport http.RoundTripper
	config    RetryConfig
}

// RoundTrip implements http.RoundTripper by adding retry logic with exponential backoff.
// It preserves the request body across retries by caching it in memory.
// The backoff duration increases exponentially: backoff * (attempt + 1).
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	var bodyBytes []byte

	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		_ = req.Body.Close()
	}

	for attempt := 0; attempt <= t.config.MaxRetries; attempt++ {
		if bodyBytes != nil {
			req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
		}

		resp, err = t.transport.RoundTrip(req)

		// Check if we should retry
		shouldRetry := t.config.RetryIf != nil && t.config.RetryIf(resp, err)

		// If this is not the last attempt and we should retry, wait and continue
		if attempt < t.config.MaxRetries && shouldRetry {
			// Exponential backoff: backoff * (attempt + 1)
			time.Sleep(t.config.Backoff * time.Duration(attempt+1))
			continue
		}

		// Either last attempt or shouldn't retry, so return
		break
	}

	return resp, err
}

// loggingTransport is an http.RoundTripper that logs HTTP requests and responses.
// It logs the request method, URL, response status code, and timing information
// using the configured logger.
type loggingTransport struct {
	transport http.RoundTripper
	logger    *log.Logger
}

// RoundTrip implements http.RoundTripper by logging the request/response and delegating to the wrapped transport.
func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	t.logger.Printf("→ %s %s", req.Method, req.URL.String())

	resp, err := t.transport.RoundTrip(req)
	duration := time.Since(start)
	if err != nil {
		t.logger.Printf("← Error: %v (took %v)", err, duration)
	} else {

		t.logger.Printf("← %d (took %v)", resp.StatusCode, duration)
	}

	return resp, err
}

//TODO add metrics transport
//TODO add rate limiting transport
