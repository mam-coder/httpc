// Package httpc provides tests for automatic retry logic.
// This file contains tests for retry configuration, exponential backoff,
// custom retry conditions, and retry behavior with different status codes.
package httpc

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries=3, got %d", config.MaxRetries)
	}

	if config.Backoff != time.Second {
		t.Errorf("Expected Backoff=1s, got %v", config.Backoff)
	}

	if config.RetryIf == nil {
		t.Error("Expected RetryIf function to be set")
	}
}

func TestDefaultRetryCondition(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		err        error
		want       bool
	}{
		{"Error should retry", 0, errors.New("network error"), true},
		{"500 should retry", 500, nil, true},
		{"502 should retry", 502, nil, true},
		{"503 should retry", 503, nil, true},
		{"429 should retry", 429, nil, true},
		{"200 should not retry", 200, nil, false},
		{"404 should not retry", 404, nil, false},
		{"400 should not retry", 400, nil, false},
		{"201 should not retry", 201, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resp *http.Response
			if tt.statusCode > 0 {
				resp = &http.Response{StatusCode: tt.statusCode}
			}

			got := defaultRetryCondition(resp, tt.err)
			if got != tt.want {
				t.Errorf("defaultRetryCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_WithRetry_Success(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := RetryConfig{
		MaxRetries: 3,
		Backoff:    10 * time.Millisecond,
		RetryIf:    defaultRetryCondition,
	}

	client := NewClient(WithRetry(config))
	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("Get() with retry failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Should only call once since it succeeded
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestClient_WithRetry_EventualSuccess(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := callCount.Add(1)
		if count < 3 {
			// Fail first 2 attempts
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			// Succeed on 3rd attempt
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	config := RetryConfig{
		MaxRetries: 3,
		Backoff:    10 * time.Millisecond,
		RetryIf:    defaultRetryCondition,
	}

	client := NewClient(WithRetry(config))
	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("Get() with retry failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Should have retried twice and succeeded on third attempt
	if callCount.Load() != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount.Load())
	}
}

func TestClient_WithRetry_MaxRetriesExceeded(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		// Always fail
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := RetryConfig{
		MaxRetries: 2,
		Backoff:    10 * time.Millisecond,
		RetryIf:    defaultRetryCondition,
	}

	client := NewClient(WithRetry(config))
	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	// Should still get a response even after max retries
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}

	// Should call: initial + 2 retries = 3 total
	if callCount.Load() != 3 {
		t.Errorf("Expected 3 calls (1 initial + 2 retries), got %d", callCount.Load())
	}
}

func TestClient_WithRetry_NoRetryOn4xx(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	config := RetryConfig{
		MaxRetries: 3,
		Backoff:    10 * time.Millisecond,
		RetryIf:    defaultRetryCondition,
	}

	client := NewClient(WithRetry(config))
	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}

	// Should not retry 404
	if callCount != 1 {
		t.Errorf("Expected 1 call (no retries), got %d", callCount)
	}
}

func TestClient_WithRetry_RetryOn429(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := callCount.Add(1)
		if count == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	config := RetryConfig{
		MaxRetries: 2,
		Backoff:    10 * time.Millisecond,
		RetryIf:    defaultRetryCondition,
	}

	client := NewClient(WithRetry(config))
	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Should retry once after 429
	if callCount.Load() != 2 {
		t.Errorf("Expected 2 calls, got %d", callCount.Load())
	}
}

func TestClient_WithRetry_CustomRetryCondition(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	// Custom retry condition that retries on 400
	config := RetryConfig{
		MaxRetries: 2,
		Backoff:    10 * time.Millisecond,
		RetryIf: func(resp *http.Response, err error) bool {
			return resp != nil && resp.StatusCode == 400
		},
	}

	client := NewClient(WithRetry(config))
	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	// Should retry on 400 with custom condition
	if callCount.Load() != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount.Load())
	}
}

func TestClient_WithRetry_BackoffTiming(t *testing.T) {
	var callCount atomic.Int32
	callTimes := make([]time.Time, 0)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		callTimes = append(callTimes, time.Now())
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	backoff := 50 * time.Millisecond
	config := RetryConfig{
		MaxRetries: 2,
		Backoff:    backoff,
		RetryIf:    defaultRetryCondition,
	}

	client := NewClient(WithRetry(config))

	start := time.Now()
	_, _ = client.Get(server.URL)
	elapsed := time.Since(start)

	// Should have: initial call + backoff*1 + call + backoff*2 + call
	// Minimum time: 50ms + 100ms = 150ms
	minExpected := backoff + backoff*2

	if elapsed < minExpected {
		t.Errorf("Expected at least %v elapsed, got %v", minExpected, elapsed)
	}

	if callCount.Load() != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount.Load())
	}
}

func TestClient_WithoutRetry(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Client without retry configuration
	client := NewClient()
	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}

	// Should only call once (no retries)
	if callCount != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
}

func TestRetryConfig_ZeroBackoff(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Test with zero backoff (immediate retries)
	config := RetryConfig{
		MaxRetries: 2,
		Backoff:    0,
		RetryIf:    defaultRetryCondition,
	}

	client := NewClient(WithRetry(config))

	start := time.Now()
	_, _ = client.Get(server.URL)
	elapsed := time.Since(start)

	// Should be very fast with no backoff
	if elapsed > 100*time.Millisecond {
		t.Errorf("Expected fast completion with zero backoff, took %v", elapsed)
	}

	if callCount.Load() != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount.Load())
	}
}

// This test is no longer applicable since interceptors are now transports
// and don't return errors in the same way. Removing it.
