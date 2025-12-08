// Package httpc provides tests for HTTP client functionality.
// This file contains comprehensive tests for the DebugTransport implementation
// including request/response logging, sensitive header masking, and timing.
package httpc

import (
	"bytes"
	"compress/gzip"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDebugTransport_Creation(t *testing.T) {
	dt := NewDebugTransport(nil, true)

	if dt == nil {
		t.Fatal("NewDebugTransport returned nil")
	}

	if !dt.Debug {
		t.Error("Expected Debug to be true")
	}

	if dt.Logger == nil {
		t.Error("Expected Logger to be set")
	}

	if !dt.LogBody {
		t.Error("Expected LogBody to be true by default")
	}

	if dt.MaxBodySize != 1024*1024 {
		t.Errorf("Expected MaxBodySize to be 1MB, got %d", dt.MaxBodySize)
	}
}

func TestDebugTransport_Disabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dt := NewDebugTransport(nil, false)
	req, _ := http.NewRequest("GET", server.URL, nil)

	resp, err := dt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestDebugTransport_LogsRequest(t *testing.T) {
	var logBuf bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dt := NewDebugTransport(nil, true)
	dt.Logger = log.New(&logBuf, "[TEST] ", 0)

	req, _ := http.NewRequest("GET", server.URL, nil)
	_, err := dt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}

	logOutput := logBuf.String()

	// Check that the request was logged
	if !strings.Contains(logOutput, "GET") {
		t.Error("Expected log to contain GET method")
	}

	if !strings.Contains(logOutput, server.URL) {
		t.Error("Expected log to contain request URL")
	}

	// Check that response was logged
	if !strings.Contains(logOutput, "200") {
		t.Error("Expected log to contain status code 200")
	}
}

func TestDebugTransport_LogsHeaders(t *testing.T) {
	var logBuf bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Response", "test-value")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dt := NewDebugTransport(nil, true)
	dt.Logger = log.New(&logBuf, "[TEST] ", 0)

	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("X-Custom-Request", "test-request")

	_, err := dt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}

	logOutput := logBuf.String()

	// Check request headers were logged
	if !strings.Contains(logOutput, "X-Custom-Request") {
		t.Error("Expected log to contain request header")
	}

	// Check response headers were logged
	if !strings.Contains(logOutput, "X-Custom-Response") {
		t.Error("Expected log to contain response header")
	}
}

func TestDebugTransport_HidesSensitiveHeaders(t *testing.T) {
	var logBuf bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dt := NewDebugTransport(nil, true)
	dt.Logger = log.New(&logBuf, "[TEST] ", 0)

	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	req.Header.Set("X-API-Key", "secret-api-key")
	req.Header.Set("Cookie", "session=secret")

	_, err := dt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}

	logOutput := logBuf.String()

	// Sensitive values should be masked
	if strings.Contains(logOutput, "secret-token") {
		t.Error("Authorization token should be masked")
	}

	if strings.Contains(logOutput, "secret-api-key") {
		t.Error("API key should be masked")
	}

	if strings.Contains(logOutput, "session=secret") {
		t.Error("Cookie should be masked")
	}

	// Should contain masked placeholder
	if !strings.Contains(logOutput, "***MODIFIED***SENSITIVE HEADER***") {
		t.Error("Expected sensitive headers to be masked")
	}
}

func TestDebugTransport_LogsBody(t *testing.T) {
	var logBuf bytes.Buffer
	requestBody := "test request body"
	responseBody := "test response body"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(responseBody))
	}))
	defer server.Close()

	dt := NewDebugTransport(nil, true)
	dt.Logger = log.New(&logBuf, "[TEST] ", 0)

	req, _ := http.NewRequest("POST", server.URL, strings.NewReader(requestBody))
	_, err := dt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}

	logOutput := logBuf.String()

	if !strings.Contains(logOutput, requestBody) {
		t.Error("Expected log to contain request body")
	}

	if !strings.Contains(logOutput, responseBody) {
		t.Error("Expected log to contain response body")
	}
}

func TestDebugTransport_WithClient(t *testing.T) {
	var logBuf bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a debug transport and inject custom logger
	dt := NewDebugTransport(nil, true)
	dt.Logger = log.New(&logBuf, "[CLIENT] ", 0)

	client := NewClient(
		WithInterceptor(func(rt http.RoundTripper) http.RoundTripper {
			return dt
		}),
	)

	_, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	logOutput := logBuf.String()

	if !strings.Contains(logOutput, "GET") {
		t.Error("Expected request to be logged")
	}

	if !strings.Contains(logOutput, "200") {
		t.Error("Expected response to be logged")
	}
}

func TestDebugTransport_LogsTiming(t *testing.T) {
	var logBuf bytes.Buffer
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dt := NewDebugTransport(nil, true)
	dt.Logger = log.New(&logBuf, "[TEST] ", 0)

	req, _ := http.NewRequest("GET", server.URL, nil)
	_, err := dt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}

	logOutput := logBuf.String()

	// Should log the duration
	if !strings.Contains(logOutput, "took") {
		t.Error("Expected log to contain timing information")
	}
}

func TestDebugTransport_HandlesErrors(t *testing.T) {
	var logBuf bytes.Buffer

	dt := NewDebugTransport(nil, true)
	dt.Logger = log.New(&logBuf, "[TEST] ", 0)

	// Request to invalid URL
	req, _ := http.NewRequest("GET", "http://invalid-domain-12345.local", nil)
	_, err := dt.RoundTrip(req)

	if err == nil {
		t.Error("Expected error for invalid domain")
	}

	logOutput := logBuf.String()

	// Should log the error
	if !strings.Contains(logOutput, "Error") {
		t.Error("Expected error to be logged")
	}
}

func TestDebugTransport_IsSensitive(t *testing.T) {
	dt := NewDebugTransport(nil, true)

	tests := []struct {
		header   string
		expected bool
	}{
		{"Authorization", true},
		{"authorization", true},
		{"X-API-Key", true},
		{"x-api-key", true},
		{"api-key", true},
		{"Cookie", true},
		{"cookie", true},
		{"Content-Type", false},
		{"Accept", false},
		{"User-Agent", false},
		{"X-Custom-Header", false},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			got := dt.isSensitive(tt.header)
			if got != tt.expected {
				t.Errorf("isSensitive(%q) = %v, want %v", tt.header, got, tt.expected)
			}
		})
	}
}

func TestDebugTransport_GzipDecoding(t *testing.T) {
	var logBuf bytes.Buffer
	responseBody := "This is a gzipped response body"

	// Create a test server that returns gzipped content
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Compress the response
		w.Header().Set("Content-Encoding", "gzip")
		gzWriter := gzip.NewWriter(w)
		gzWriter.Write([]byte(responseBody))
		gzWriter.Close()
	}))
	defer server.Close()

	dt := NewDebugTransport(nil, true)
	dt.Logger = log.New(&logBuf, "[TEST] ", 0)

	req, _ := http.NewRequest("GET", server.URL, nil)
	_, err := dt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}

	logOutput := logBuf.String()

	// The log should contain the decoded (uncompressed) body
	if !strings.Contains(logOutput, responseBody) {
		t.Errorf("Expected log to contain decoded body %q, got: %s", responseBody, logOutput)
	}

	// Should not contain gzip header bytes
	if strings.Contains(logOutput, "\x1f\x8b") {
		t.Error("Log contains gzip header, body was not decoded")
	}
}
