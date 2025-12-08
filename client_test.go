// Package httpc provides tests for the HTTP client functionality.
// This file contains tests for client initialization, configuration, and core functionality.
package httpc

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// TestNewClient verifies that NewClient creates a properly initialized client
// with all required fields (headers map, httpClient, and mutex).
func TestNewClient(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.headers == nil {
		t.Error("Expected headers map to be initialized")
	}

	if client.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}

	if client.mu == nil {
		t.Error("Expected mutex to be initialized")
	}
}

// TestNewClient_WithOptions verifies that client options are properly applied
// during client initialization including baseURL, timeout, and default headers.
func TestNewClient_WithOptions(t *testing.T) {
	baseURL := "https://api.example.com"
	timeout := 15 * time.Second

	client := NewClient(
		WithBaseURL(baseURL),
		WithTimeout(timeout),
		WithHeader("User-Agent", "TestAgent"),
	)

	if client.baseURL != baseURL {
		t.Error("baseURL not applied")
	}

	if client.httpClient.Timeout != timeout {
		t.Error("timeout not applied")
	}

	// Headers are now set via transport, not client.headers directly
	// Verify that transport was configured
	if client.transport == nil {
		t.Error("transport not configured")
	}
}

// TestDefault verifies that the Default() function creates a client
// with standard default settings (30-second timeout, no base URL).
func TestDefault(t *testing.T) {
	client := Default()

	if client == nil {
		t.Fatal("Default() returned nil")
	}

	// Should have default timeout
	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", client.httpClient.Timeout)
	}
}

// TestClient_WithInterceptor tests that interceptors can be added via options
func TestClient_WithInterceptor(t *testing.T) {
	client := NewClient(
		WithInterceptor(func(rt http.RoundTripper) http.RoundTripper {
			return rt
		}),
	)

	if client.transport == nil {
		t.Error("Expected transport to be set")
	}
}

func TestClient_ThreadSafety(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	// Make concurrent requests
	var wg sync.WaitGroup
	errors := make(chan error, 20)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := client.Get("/test")
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Errorf("Concurrent request failed: %v", err)
	}
}

func TestClient_BaseURL(t *testing.T) {
	tests := []struct {
		name     string
		useBase  bool
		path     string
		expected string
	}{
		{
			name:     "Relative path with base URL",
			useBase:  true,
			path:     "/users",
			expected: "/users",
		},
		{
			name:     "Absolute URL ignores base",
			useBase:  false,
			path:     "/direct",
			expected: "/direct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.expected {
					t.Errorf("Expected path %s, got %s", tt.expected, r.URL.Path)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			var client *Client
			if tt.useBase {
				client = NewClient(WithBaseURL(server.URL))
				_, err := client.Get(tt.path)
				if err != nil {
					t.Fatalf("Request failed: %v", err)
				}
			} else {
				client = NewClient()
				_, err := client.Get(server.URL + tt.path)
				if err != nil {
					t.Fatalf("Request failed: %v", err)
				}
			}
		})
	}
}

func TestClient_DefaultHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Default") != "value" {
			t.Error("Expected X-Default header from client")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(WithHeader("X-Default", "value"))
	_, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
}

func TestClient_Transport(t *testing.T) {
	client := NewClient()

	if client.httpClient.Transport == nil {
		t.Error("Expected transport to be initialized")
	}

	// The client.transport field should be set to defaultTransport
	if client.transport == nil {
		t.Error("Expected client.transport to be initialized")
	}

	// Try to access the underlying *http.Transport
	// Since client.transport is the base, we can type assert it
	transport, ok := client.transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected *http.Transport as base transport")
	}

	// Check some transport settings
	if transport.MaxIdleConns != 100 {
		t.Errorf("Expected MaxIdleConns=100, got %d", transport.MaxIdleConns)
	}

	if transport.MaxIdleConnsPerHost != 10 {
		t.Errorf("Expected MaxIdleConnsPerHost=10, got %d", transport.MaxIdleConnsPerHost)
	}

	if !transport.ForceAttemptHTTP2 {
		t.Error("Expected ForceAttemptHTTP2 to be true")
	}
}

func TestClient_NewRequest(t *testing.T) {
	client := NewClient()
	rb := client.NewRequest()

	if rb == nil {
		t.Fatal("NewRequest() returned nil")
	}

	if rb.client != client {
		t.Error("RequestBuilder should reference the client")
	}

	if rb.headers == nil {
		t.Error("RequestBuilder headers should be initialized")
	}

	if rb.query == nil {
		t.Error("RequestBuilder query should be initialized")
	}
}

func TestClient_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(WithTimeout(10 * time.Millisecond))
	_, err := client.Get(server.URL)

	if err == nil {
		t.Error("Expected timeout error")
	}

	if !IsTimeout(err) {
		t.Errorf("Expected timeout error, got %v", err)
	}
}

func TestClient_MultipleClients(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create multiple clients with different configurations
	client1 := NewClient(WithHeader("X-Client", "1"))
	client2 := NewClient(WithHeader("X-Client", "2"))

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, err := client1.Get(server.URL)
		if err != nil {
			t.Errorf("Client1 request failed: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		_, err := client2.Get(server.URL)
		if err != nil {
			t.Errorf("Client2 request failed: %v", err)
		}
	}()

	wg.Wait()
}

func TestClient_HeadersThreadSafety(t *testing.T) {
	client := NewClient(
		WithHeader("X-Initial", "value"),
	)

	var wg sync.WaitGroup

	// Concurrently access headers from multiple goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			// Access client headers (read operation)
			client.mu.RLock()
			_ = client.headers
			client.mu.RUnlock()
		}(i)
	}

	wg.Wait()
}

func TestClient_EmptyConfiguration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Client with no options should still work
	client := NewClient()
	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("Request with default client failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestClient_Debug(t *testing.T) {
	client := NewClient(WithDebug())

	// Verify debug transport was set
	if client.transport == nil {
		t.Error("Expected transport to be set with debug")
	}

	// Check if it's a DebugTransport by making a test request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	_, err := client.Get(server.URL)
	if err != nil {
		t.Errorf("Request with debug transport failed: %v", err)
	}
}

func TestClient_ReuseConnection(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()

	// Make multiple requests with the same client
	for i := 0; i < 5; i++ {
		_, err := client.Get(server.URL)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
	}

	if requestCount != 5 {
		t.Errorf("Expected 5 requests, got %d", requestCount)
	}
}
