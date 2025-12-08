// Package httpc provides tests for request builder functionality.
// This file contains tests for the fluent RequestBuilder API including
// URL resolution, query parameters, headers, context, and error handling.
package httpc

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRequestBuilder_Method(t *testing.T) {
	client := NewClient()
	rb := client.NewRequest()

	rb.Method("POST")

	if rb.method != "POST" {
		t.Errorf("Expected method POST, got %s", rb.method)
	}
}

func TestRequestBuilder_URL(t *testing.T) {
	client := NewClient()
	rb := client.NewRequest()

	rb.URL("/api/users")

	if rb.url != "/api/users" {
		t.Errorf("Expected URL /api/users, got %s", rb.url)
	}
}

func TestRequestBuilder_Header(t *testing.T) {
	client := NewClient()
	rb := client.NewRequest()

	rb.Header("Content-Type", "application/json").
		Header("Authorization", "Bearer token")

	if rb.headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type header to be application/json")
	}

	if rb.headers["Authorization"] != "Bearer token" {
		t.Errorf("Expected Authorization header to be Bearer token")
	}
}

func TestRequestBuilder_Query(t *testing.T) {
	client := NewClient()
	rb := client.NewRequest()

	rb.Query("page", "1").
		Query("limit", "10").
		Query("tag", "golang").
		Query("tag", "http")

	if rb.query.Get("page") != "1" {
		t.Errorf("Expected page=1")
	}

	if rb.query.Get("limit") != "10" {
		t.Errorf("Expected limit=10")
	}

	// Check multiple values for same key
	tags := rb.query["tag"]
	if len(tags) != 2 {
		t.Errorf("Expected 2 tag values, got %d", len(tags))
	}
}

func TestRequestBuilder_QueryParams(t *testing.T) {
	client := NewClient()
	rb := client.NewRequest()

	params := map[string]string{
		"page":  "1",
		"limit": "10",
		"sort":  "name",
	}

	rb.QueryParams(params)

	for key, value := range params {
		if rb.query.Get(key) != value {
			t.Errorf("Expected %s=%s, got %s", key, value, rb.query.Get(key))
		}
	}
}

func TestRequestBuilder_Body(t *testing.T) {
	client := NewClient()
	rb := client.NewRequest()

	body := strings.NewReader("test body")
	rb.Body(body)

	if rb.body != body {
		t.Error("Expected body to be set")
	}
}

func TestRequestBuilder_Timeout(t *testing.T) {
	client := NewClient()
	rb := client.NewRequest()

	timeout := 5 * time.Second
	rb.Timeout(timeout)

	if rb.timeout != timeout {
		t.Errorf("Expected timeout %v, got %v", timeout, rb.timeout)
	}
}

func TestRequestBuilder_Timeout_ActuallyWorks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()

	// Request should timeout
	_, err := client.NewRequest().
		Method("GET").
		URL(server.URL).
		Timeout(50 * time.Millisecond).
		Do()

	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}

	if !IsTimeout(err) {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestRequestBuilder_Context(t *testing.T) {
	client := NewClient()
	rb := client.NewRequest()

	ctx := context.WithValue(context.Background(), "key", "value")
	rb.Context(ctx)

	if rb.ctx != ctx {
		t.Error("Expected context to be set")
	}
}

func TestRequestBuilder_buildURL_NoQuery(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		url      string
		expected string
	}{
		{
			name:     "Relative URL with base",
			baseURL:  "https://api.example.com",
			url:      "/users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "Absolute URL ignores base",
			baseURL:  "https://api.example.com",
			url:      "https://other.com/users",
			expected: "https://other.com/users",
		},
		{
			name:     "Empty base URL",
			baseURL:  "",
			url:      "/users",
			expected: "/users",
		},
		{
			name:     "Base URL with trailing slash",
			baseURL:  "https://api.example.com/",
			url:      "/users",
			expected: "https://api.example.com/users",
		},
		{
			name:     "URL without leading slash",
			baseURL:  "https://api.example.com",
			url:      "users",
			expected: "https://api.example.com/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(WithBaseURL(tt.baseURL))
			rb := client.NewRequest().URL(tt.url)

			got := rb.buildURL()
			if got != tt.expected {
				t.Errorf("buildURL() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRequestBuilder_buildURL_WithQuery(t *testing.T) {
	client := NewClient(WithBaseURL("https://api.example.com"))
	rb := client.NewRequest().
		URL("/users").
		Query("page", "1").
		Query("limit", "10")

	got := rb.buildURL()

	// Check if URL contains query parameters (order may vary)
	if !strings.Contains(got, "page=1") {
		t.Error("Expected URL to contain page=1")
	}
	if !strings.Contains(got, "limit=10") {
		t.Error("Expected URL to contain limit=10")
	}
	if !strings.HasPrefix(got, "https://api.example.com/users?") {
		t.Errorf("Expected URL to start with https://api.example.com/users?, got %s", got)
	}
}

func TestRequestBuilder_isAbsoluteURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"https://example.com", true},
		{"http://example.com", true},
		{"/api/users", false},
		{"api/users", false},
		{"", false},
		{"ftp://example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := isAbsoluteURL(tt.url); got != tt.want {
				t.Errorf("isAbsoluteURL(%s) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestRequestBuilder_ChainedCalls(t *testing.T) {
	client := NewClient(WithBaseURL("https://api.example.com"))

	rb := client.NewRequest().
		Method("POST").
		URL("/users").
		Header("Content-Type", "application/json").
		Header("Authorization", "Bearer token").
		Query("notify", "true").
		Body(strings.NewReader(`{"name":"John"}`)).
		Timeout(10 * time.Second)

	if rb.method != "POST" {
		t.Error("Method not set correctly")
	}
	if rb.url != "/users" {
		t.Error("URL not set correctly")
	}
	if len(rb.headers) != 2 {
		t.Errorf("Expected 2 headers, got %d", len(rb.headers))
	}
	if rb.query.Get("notify") != "true" {
		t.Error("Query param not set correctly")
	}
	if rb.body == nil {
		t.Error("Body not set")
	}
	if rb.timeout != 10*time.Second {
		t.Error("Timeout not set correctly")
	}
}

func TestRequestBuilder_resolveURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		url      string
		expected string
	}{
		{
			name:     "Both empty",
			baseURL:  "",
			url:      "",
			expected: "",
		},
		{
			name:     "Only base URL",
			baseURL:  "https://api.example.com",
			url:      "",
			expected: "https://api.example.com",
		},
		{
			name:     "Base with path",
			baseURL:  "https://api.example.com/v1",
			url:      "/users",
			expected: "https://api.example.com/v1/users",
		},
		{
			name:     "Spaces trimmed",
			baseURL:  "  https://api.example.com  ",
			url:      "  /users  ",
			expected: "https://api.example.com/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(WithBaseURL(tt.baseURL))
			rb := client.NewRequest().URL(tt.url)

			got := rb.resolveURL()
			if got != tt.expected {
				t.Errorf("resolveURL() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRequestBuilder_applyHeaders(t *testing.T) {
	client := NewClient(
		WithBaseURL("https://api.example.com"),
		WithHeader("X-Client-Header", "client-value"),
	)

	rb := client.NewRequest().
		Header("X-Request-Header", "request-value")

	// Since applyHeaders is called during Do(), we'll test it indirectly
	// by checking the headers map
	if rb.headers["X-Request-Header"] != "request-value" {
		t.Error("Request-specific header not set")
	}

	// Client headers are now set via transport, not directly in client.headers
	// Just verify the headers map is initialized
	if client.headers == nil {
		t.Error("Client headers map should be initialized")
	}
}

func TestRequestBuilder_ErrorHandling(t *testing.T) {
	client := NewClient()

	// Simulate an error during building
	rb := client.NewRequest()
	rb.err = io.ErrUnexpectedEOF

	_, err := rb.Do()
	if err != io.ErrUnexpectedEOF {
		t.Errorf("Expected error to be propagated, got %v", err)
	}
}

func TestRequestBuilder_DefaultContext(t *testing.T) {
	client := NewClient()
	rb := client.NewRequest()

	if rb.ctx == nil {
		t.Error("Expected default context to be set")
	}

	if rb.ctx != context.Background() {
		t.Error("Expected default context to be context.Background()")
	}
}

func TestRequestBuilder_QueryWithExistingParams(t *testing.T) {
	client := NewClient(WithBaseURL("https://api.example.com"))

	// URL already has query parameters
	rb := client.NewRequest().
		URL("/users?existing=param").
		Query("new", "value")

	url := rb.buildURL()

	if !strings.Contains(url, "existing=param") {
		t.Error("Existing query param should be preserved")
	}
	if !strings.Contains(url, "new=value") {
		t.Error("New query param should be added")
	}
}

func TestRequestBuilder_Query_NilInitialization(t *testing.T) {
	client := NewClient()
	rb := client.NewRequest()

	// Set query to nil then use Query method
	rb.query = nil
	rb.Query("key", "value")

	if rb.query.Get("key") != "value" {
		t.Error("Query should initialize nil query map")
	}
}

func TestRequestBuilder_QueryParams_NilInitialization(t *testing.T) {
	client := NewClient()
	rb := client.NewRequest()

	// Set query to nil then use QueryParams method
	rb.query = nil
	rb.QueryParams(map[string]string{"key": "value"})

	if rb.query.Get("key") != "value" {
		t.Error("QueryParams should initialize nil query map")
	}
}

func TestRequestBuilder_buildURL_InvalidURL(t *testing.T) {
	client := NewClient()
	rb := client.NewRequest()

	// Invalid URL that can't be parsed
	rb.url = "://invalid"
	rb.Query("test", "value")

	// Should return the URL as-is without adding query params
	result := rb.buildURL()
	if result != "://invalid" {
		t.Errorf("Expected invalid URL to be returned as-is, got %s", result)
	}
}

func TestRequestBuilder_Do_WithNetworkError(t *testing.T) {
	client := NewClient()

	rb := client.NewRequest().
		Method("GET").
		URL("http://invalid-domain-that-does-not-exist-12345.com")

	_, err := rb.Do()
	if err == nil {
		t.Error("Expected network error for invalid domain")
	}
}
