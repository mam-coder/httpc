// Package httpc provides tests for HTTP request interceptors (now transport-based).
// This file contains tests for transport wrappers including authentication,
// logging, rate limiting, validation, and custom transports.
package httpc

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHeaderTransport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers were set
		if r.Header.Get("X-Custom") != "value" {
			t.Errorf("Expected X-Custom header, got %q", r.Header.Get("X-Custom"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(
		WithHeader("X-Custom", "value"),
	)

	_, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
}

func TestMultipleHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Header-1") != "value1" {
			t.Error("Expected X-Header-1 to be value1")
		}
		if r.Header.Get("X-Header-2") != "value2" {
			t.Error("Expected X-Header-2 to be value2")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(
		WithHeader("X-Header-1", "value1"),
		WithHeader("X-Header-2", "value2"),
	)

	_, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
}

func TestBaseAuthTransport(t *testing.T) {
	username := "testuser"
	password := "testpass"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUser, gotPass, ok := r.BasicAuth()
		if !ok {
			t.Error("Expected basic auth to be set")
		}
		if gotUser != username {
			t.Errorf("Expected username %q, got %q", username, gotUser)
		}
		if gotPass != password {
			t.Errorf("Expected password %q, got %q", password, gotPass)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(
		WithBaseAuth(username, password),
	)

	_, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
}

func TestAuthTransport(t *testing.T) {
	token := "test-token-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		expected := "Bearer " + token
		if auth != expected {
			t.Errorf("Expected Authorization %q, got %q", expected, auth)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(
		WithAuthorization(token),
	)

	_, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
}

func TestBlockListTransport(t *testing.T) {
	tests := []struct {
		name        string
		blockedList []string
		url         string
		shouldBlock bool
	}{
		{
			name:        "Blocked domain",
			blockedList: []string{"blocked.com"},
			url:         "http://blocked.com/path",
			shouldBlock: true,
		},
		{
			name:        "Allowed domain",
			blockedList: []string{"blocked.com"},
			url:         "http://allowed.com/path",
			shouldBlock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.shouldBlock {
					t.Error("Server should not be reached for blocked domain")
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := NewClient(
				WithBlockedList(tt.blockedList),
			)

			// Use the test URL if it should block, otherwise use server URL
			requestURL := tt.url
			if !tt.shouldBlock {
				requestURL = server.URL
			}

			_, err := client.Get(requestURL)

			if tt.shouldBlock {
				if err == nil {
					t.Error("Expected error for blocked domain, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for allowed domain, got %v", err)
				}
			}
		})
	}
}

func TestLoggingTransport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(
		WithLogger(log.Default()), // Use default logger
	)

	// This should log the request (output to default logger)
	_, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request with logging failed: %v", err)
	}
}

func TestUserAgentOption(t *testing.T) {
	userAgent := "TestAgent/1.0"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("User-Agent")
		if got != userAgent {
			t.Errorf("Expected User-Agent %q, got %q", userAgent, got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(
		WithUserAgent(userAgent),
	)

	_, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
}

func TestContentTypeOption(t *testing.T) {
	contentType := ContentTypeJSON

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("Content-Type")
		if got != contentType {
			t.Errorf("Expected Content-Type %q, got %q", contentType, got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(
		WithContentType(contentType),
	)

	_, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
}

func TestAcceptOption(t *testing.T) {
	accept := ContentTypeJSON

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("Accept")
		if got != accept {
			t.Errorf("Expected Accept %q, got %q", accept, got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(
		WithAccept(accept),
	)

	_, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
}

func TestApiKeyOption(t *testing.T) {
	apiKey := "secret-key-123"
	headerName := "X-API-Key"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get(headerName)
		if got != apiKey {
			t.Errorf("Expected %s %q, got %q", headerName, apiKey, got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(
		WithApiKey(headerName, apiKey),
	)

	_, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
}

func TestApiKeyOption_DefaultHeader(t *testing.T) {
	apiKey := "secret-key-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("X-Api-Key")
		if got != apiKey {
			t.Errorf("Expected X-Api-Key %q, got %q", apiKey, got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(
		WithApiKey("", apiKey), // Empty headerName should use default
	)

	_, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
}

func TestRequestIdOption(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("X-Request-Id")
		if got == "" {
			t.Error("Expected X-Request-Id to be set")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(
		WithRequestId("X-Request-Id"),
	)

	_, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
}

func TestRequestIdOption_DefaultHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("X-Request-Id")
		if got == "" {
			t.Error("Expected X-Request-Id to be set")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(
		WithRequestId(""), // Empty headerName should use default
	)

	_, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
}

func TestChainedInterceptors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify all headers were set in order
		if r.Header.Get("X-Header-1") != "value1" {
			t.Error("X-Header-1 not set")
		}
		if r.Header.Get("X-Header-2") != "value2" {
			t.Error("X-Header-2 not set")
		}
		if r.Header.Get("User-Agent") != "TestAgent" {
			t.Error("User-Agent not set")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(
		WithHeader("X-Header-1", "value1"),
		WithHeader("X-Header-2", "value2"),
		WithUserAgent("TestAgent"),
	)

	_, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
}
