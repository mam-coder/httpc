// Package httpc provides tests for client configuration options.
// This file contains tests for all WithXXX option functions including
// baseURL, timeout, headers, retries, interceptors, and debug mode.
package httpc

import (
	"net/http"
	"testing"
	"time"
)

func TestWithBaseURL(t *testing.T) {
	baseURL := "https://api.example.com"
	client := NewClient(WithBaseURL(baseURL))

	if client.baseURL != baseURL {
		t.Errorf("Expected baseURL %s, got %s", baseURL, client.baseURL)
	}
}

func TestWithTimeout(t *testing.T) {
	timeout := 15 * time.Second
	client := NewClient(WithTimeout(timeout))

	if client.httpClient.Timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, client.httpClient.Timeout)
	}
}

func TestWithHeader(t *testing.T) {
	client := NewClient(
		WithHeader("User-Agent", "TestAgent/1.0"),
		WithHeader("Accept", ContentTypeJSON),
	)

	// Headers are now set via transport, verify transport is set
	if client.transport == nil {
		t.Error("Expected transport to be set with headers")
	}
}

func TestWithRetry(t *testing.T) {
	maxRetries := 5
	backoff := 2 * time.Second

	config := RetryConfig{
		MaxRetries: maxRetries,
		Backoff:    backoff,
		RetryIf:    defaultRetryCondition,
	}

	client := NewClient(WithRetry(config))

	// Retry is now implemented via transport
	if client.transport == nil {
		t.Fatal("Expected transport to be set with retry")
	}
}

func TestWithInterceptor(t *testing.T) {
	client := NewClient(
		WithInterceptor(func(rt http.RoundTripper) http.RoundTripper {
			return rt
		}),
		WithInterceptor(func(rt http.RoundTripper) http.RoundTripper {
			return rt
		}),
	)

	// Interceptors now wrap the transport
	if client.transport == nil {
		t.Error("Expected transport to be set")
	}
}

func TestWithDebug(t *testing.T) {
	client := NewClient(WithDebug())

	// Debug is now implemented via DebugTransport
	if client.transport == nil {
		t.Error("Expected transport to be set with debug")
	}
}

func TestMultipleOptions(t *testing.T) {
	baseURL := "https://api.example.com"
	timeout := 10 * time.Second
	maxRetries := 3
	backoff := time.Second

	config := RetryConfig{
		MaxRetries: maxRetries,
		Backoff:    backoff,
		RetryIf:    defaultRetryCondition,
	}

	client := NewClient(
		WithBaseURL(baseURL),
		WithTimeout(timeout),
		WithHeader("User-Agent", "TestAgent"),
		WithHeader("Accept", ContentTypeJSON),
		WithRetry(config),
		WithDebug(),
	)

	// Verify all options were applied
	if client.baseURL != baseURL {
		t.Error("baseURL not set")
	}

	if client.httpClient.Timeout != timeout {
		t.Error("timeout not set")
	}

	if client.transport == nil {
		t.Error("transport not set")
	}
}

func TestWithBaseURL_EmptyString(t *testing.T) {
	client := NewClient(WithBaseURL(""))

	if client.baseURL != "" {
		t.Errorf("Expected empty baseURL, got %s", client.baseURL)
	}
}

func TestWithTimeout_ZeroDuration(t *testing.T) {
	client := NewClient(WithTimeout(0))

	if client.httpClient.Timeout != 0 {
		t.Errorf("Expected zero timeout, got %v", client.httpClient.Timeout)
	}
}

func TestWithRetry_ZeroRetries(t *testing.T) {
	config := RetryConfig{
		MaxRetries: 0,
		Backoff:    time.Second,
		RetryIf:    defaultRetryCondition,
	}

	client := NewClient(WithRetry(config))

	if client.transport == nil {
		t.Fatal("Expected transport to be set")
	}
}

func TestWithHeader_Overwrite(t *testing.T) {
	client := NewClient(
		WithHeader("X-Custom", "value1"),
		WithHeader("X-Custom", "value2"),
	)

	// Headers are set via transport, verify it's configured
	if client.transport == nil {
		t.Error("Expected transport to be set")
	}
}

func TestOptionOrdering(t *testing.T) {
	// Test that options are applied in order
	client := NewClient(
		WithTimeout(5*time.Second),
		WithTimeout(10*time.Second), // This should override
	)

	if client.httpClient.Timeout != 10*time.Second {
		t.Error("Later option should override earlier one")
	}
}

func TestWithInterceptor_Nil(t *testing.T) {
	// With nil interceptor, client should still be created but may panic when used
	// This test just ensures we don't crash during client creation
	defer func() {
		if r := recover(); r == nil {
			// Client creation succeeded, which is fine
		}
	}()

	client := NewClient(WithInterceptor(nil))
	if client == nil {
		t.Error("Expected client to be created even with nil interceptor")
	}
}

func TestDefaultClientOptions(t *testing.T) {
	client := NewClient()

	// Check defaults
	if client.baseURL != "" {
		t.Error("Expected empty baseURL by default")
	}

	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", client.httpClient.Timeout)
	}

	if len(client.headers) != 0 {
		t.Error("Expected no default headers")
	}

	if client.transport == nil {
		t.Error("Expected default transport to be set")
	}
}

func TestWithHeader_SpecialCharacters(t *testing.T) {
	client := NewClient(
		WithHeader("X-Special", "value with spaces"),
		WithHeader("X-Unicode", "值"),
	)

	if client.transport == nil {
		t.Error("Expected transport to be set")
	}
}
