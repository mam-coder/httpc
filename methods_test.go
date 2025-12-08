// Package httpc provides tests for HTTP method convenience functions.
// This file contains tests for GET, POST, PUT, DELETE, PATCH methods,
// context support, and request options handling.
package httpc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	client := NewClient()
	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := resp.String()
	if body != "success" {
		t.Errorf("Expected body 'success', got %s", body)
	}
}

func TestClient_Get_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "value" {
			t.Error("Expected X-Custom header")
		}
		if r.URL.Query().Get("page") != "1" {
			t.Error("Expected page=1 query param")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	_, err := client.Get(server.URL,
		Header("X-Custom", "value"),
		WithQuery("page", "1"),
	)

	if err != nil {
		t.Fatalf("Get() with options failed: %v", err)
	}
}

func TestClient_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var data map[string]string
		json.Unmarshal(body, &data)

		if data["name"] != "John" {
			t.Error("Expected name=John in request body")
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":123}`))
	}))
	defer server.Close()

	client := NewClient()
	resp, err := client.Post(server.URL, map[string]string{"name": "John"})

	if err != nil {
		t.Fatalf("Post() failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

func TestClient_Post_NilBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	_, err := client.Post(server.URL, nil)

	if err != nil {
		t.Fatalf("Post() with nil body failed: %v", err)
	}
}

func TestClient_Put(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var data map[string]string
		json.Unmarshal(body, &data)

		if data["status"] != "active" {
			t.Error("Expected status=active in request body")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	resp, err := client.Put(server.URL, map[string]string{"status": "active"})

	if err != nil {
		t.Fatalf("Put() failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestClient_Put_NilBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	_, err := client.Put(server.URL, nil)

	if err != nil {
		t.Fatalf("Put() with nil body failed: %v", err)
	}
}

func TestClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient()
	resp, err := client.Delete(server.URL)

	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", resp.StatusCode)
	}
}

func TestClient_GetWithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if context is present
		if r.Context() == nil {
			t.Error("Expected context to be present")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	ctx := context.WithValue(context.Background(), "key", "value")
	resp, err := client.GetWithContext(ctx, server.URL)

	if err != nil {
		t.Fatalf("GetWithContext() failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestClient_GetWithContext_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.GetWithContext(ctx, server.URL)

	if err == nil {
		t.Error("Expected timeout error")
	}

	if !IsTimeout(err) {
		t.Errorf("Expected timeout error, got %v", err)
	}
}

func TestHeader_RequestOption(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-1") != "value1" {
			t.Error("Expected X-Custom-1 header")
		}
		if r.Header.Get("X-Custom-2") != "value2" {
			t.Error("Expected X-Custom-2 header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	_, err := client.Get(server.URL,
		Header("X-Custom-1", "value1"),
		Header("X-Custom-2", "value2"),
	)

	if err != nil {
		t.Fatalf("Get() with Header options failed: %v", err)
	}
}

func TestWithQuery_RequestOption(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("param1") != "value1" {
			t.Error("Expected param1=value1")
		}
		if r.URL.Query().Get("param2") != "value2" {
			t.Error("Expected param2=value2")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	_, err := client.Get(server.URL,
		WithQuery("param1", "value1"),
		WithQuery("param2", "value2"),
	)

	if err != nil {
		t.Fatalf("Get() with WithQuery options failed: %v", err)
	}
}

func TestClient_MethodsWithBaseURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/users" {
			t.Errorf("Expected path /api/users, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(WithBaseURL(server.URL))

	tests := []struct {
		name   string
		method func() (*Response, error)
	}{
		{"GET", func() (*Response, error) { return client.Get("/api/users") }},
		{"POST", func() (*Response, error) { return client.Post("/api/users", nil) }},
		{"PUT", func() (*Response, error) { return client.Put("/api/users", nil) }},
		{"DELETE", func() (*Response, error) { return client.Delete("/api/users") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.method()
			if err != nil {
				t.Fatalf("%s failed: %v", tt.name, err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}
		})
	}
}

func TestClient_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH method, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		if len(body) == 0 {
			t.Error("Expected request body")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "updated"}`))
	}))
	defer server.Close()

	client := NewClient()
	updates := map[string]string{"status": "active"}
	resp, err := client.Patch(server.URL, updates)

	if err != nil {
		t.Fatalf("Patch() failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestClient_Put_NetworkError(t *testing.T) {
	client := NewClient()

	_, err := client.Put("http://invalid-domain-that-does-not-exist-12345.com", nil)
	if err == nil {
		t.Error("Expected network error for invalid domain")
	}
}

func TestClient_Delete_NetworkError(t *testing.T) {
	client := NewClient()

	_, err := client.Delete("http://invalid-domain-that-does-not-exist-12345.com")
	if err == nil {
		t.Error("Expected network error for invalid domain")
	}
}

func TestClient_GetWithContext_NetworkError(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	_, err := client.GetWithContext(ctx, "http://invalid-domain-that-does-not-exist-12345.com")
	if err == nil {
		t.Error("Expected network error for invalid domain")
	}
}

func TestClient_Put_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT method, got %s", r.Method)
		}
		if r.Header.Get("X-Custom") != "header" {
			t.Error("Expected X-Custom header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	_, err := client.Put(server.URL, nil, Header("X-Custom", "header"))

	if err != nil {
		t.Fatalf("Put() with options failed: %v", err)
	}
}

func TestClient_Delete_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE method, got %s", r.Method)
		}
		if r.Header.Get("X-Custom") != "header" {
			t.Error("Expected X-Custom header")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient()
	_, err := client.Delete(server.URL, Header("X-Custom", "header"))

	if err != nil {
		t.Fatalf("Delete() with options failed: %v", err)
	}
}

func TestClient_GetWithContext_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "header" {
			t.Error("Expected X-Custom header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	ctx := context.Background()
	_, err := client.GetWithContext(ctx, server.URL, Header("X-Custom", "header"))

	if err != nil {
		t.Fatalf("GetWithContext() with options failed: %v", err)
	}
}
