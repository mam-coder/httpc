// Package httpc provides a modern HTTP client with fluent API and advanced features.
// This file contains the core Client type and initialization functions.
package httpc

import (
	"net/http"
	"sync"
	"time"
)

// Client is the main HTTP client type that provides a fluent API for making HTTP requests.
// It supports configuration through options, automatic retries, request interceptors,
// and various convenience methods for common HTTP operations.
//
// Client is safe for concurrent use by multiple goroutines.
type Client struct {
	baseURL    string
	headers    map[string]string
	httpClient *http.Client
	transport  http.RoundTripper
	mu         *sync.RWMutex
}

// NewClient creates a new HTTP client with the specified options.
// Options can configure the base URL, timeout, headers, retry logic,
// interceptors, and other client behavior.
//
// Example:
//
//	client := httpc.NewClient(
//	    httpc.WithBaseURL("https://api.example.com"),
//	    httpc.WithTimeout(30*time.Second),
//	    httpc.WithHeader("User-Agent", "MyApp/1.0"),
//	)
func NewClient(opts ...Option) *Client {
	client := &Client{
		headers: make(map[string]string),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		transport: defaultTransport(),
		mu:        &sync.RWMutex{},
	}

	for _, opt := range opts {
		opt(client)
	}

	client.httpClient.Transport = client.transport

	return client
}

// Default creates a client with sensible defaults.
// This is equivalent to calling NewClient() with no options.
// The returned client will have a 30-second timeout and standard transport settings.
//
// Example:
//
//	client := httpc.Default()
//	resp, err := client.Get("https://api.example.com/users")
func Default() *Client {
	return NewClient()
}
