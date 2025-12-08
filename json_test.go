// Package httpc provides tests for JSON handling functionality.
// This file contains tests for JSON marshaling/unmarshaling, convenience methods
// (GetJSON, PostJSON), and error handling for invalid JSON.
package httpc

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestBuilder_JSON(t *testing.T) {
	client := NewClient()

	data := map[string]interface{}{
		"name":  "John",
		"age":   30,
		"email": "john@example.com",
	}

	rb := client.NewRequest().
		Method("POST").
		URL("http://example.com").
		JSON(data)

	// Check that Content-Type header is set
	if rb.headers["Content-Type"] != ContentTypeJSON {
		t.Errorf("Expected Content-Type %s, got %s", ContentTypeJSON, rb.headers["Content-Type"])
	}

	// Check that body is set
	if rb.body == nil {
		t.Error("Expected body to be set")
	}
}

func TestRequestBuilder_JSON_WithError(t *testing.T) {
	client := NewClient()

	// channels cannot be marshaled to JSON
	invalidData := make(chan int)

	rb := client.NewRequest().
		Method("POST").
		URL("http://example.com").
		JSON(invalidData)

	// Check that error is set
	if rb.err == nil {
		t.Error("Expected error when marshaling invalid data")
	}
}

func TestClient_GetJSON(t *testing.T) {
	expectedData := map[string]interface{}{
		"name":  "Alice",
		"email": "alice@example.com",
		"age":   float64(25),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		w.Header().Set("Content-Type", ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedData)
	}))
	defer server.Close()

	client := NewClient()
	var result map[string]interface{}

	err := client.GetJSON(server.URL, &result)
	if err != nil {
		t.Fatalf("GetJSON() failed: %v", err)
	}

	if result["name"] != expectedData["name"] {
		t.Errorf("Expected name=%s, got %s", expectedData["name"], result["name"])
	}
	if result["email"] != expectedData["email"] {
		t.Errorf("Expected email=%s, got %s", expectedData["email"], result["email"])
	}
	if result["age"] != expectedData["age"] {
		t.Errorf("Expected age=%v, got %v", expectedData["age"], result["age"])
	}
}

func TestClient_GetJSON_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "active" {
			t.Error("Expected status=active query param")
		}
		w.Header().Set("Content-Type", ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"result": "ok"})
	}))
	defer server.Close()

	client := NewClient()
	var result map[string]string

	err := client.GetJSON(server.URL, &result, WithQuery("status", "active"))
	if err != nil {
		t.Fatalf("GetJSON() with options failed: %v", err)
	}

	if result["result"] != "ok" {
		t.Errorf("Expected result=ok, got %s", result["result"])
	}
}

func TestClient_GetJSON_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid json"))
	}))
	defer server.Close()

	client := NewClient()
	var result map[string]interface{}

	err := client.GetJSON(server.URL, &result)
	if err == nil {
		t.Error("Expected error when parsing invalid JSON")
	}
}

func TestClient_PostJSON(t *testing.T) {
	requestData := map[string]interface{}{
		"name":  "Bob",
		"email": "bob@example.com",
	}

	responseData := map[string]interface{}{
		"id":    float64(123),
		"name":  "Bob",
		"email": "bob@example.com",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Check Content-Type
		if r.Header.Get("Content-Type") != ContentTypeJSON {
			t.Errorf("Expected Content-Type %s", ContentTypeJSON)
		}

		// Read and verify request body
		body, _ := io.ReadAll(r.Body)
		var data map[string]interface{}
		json.Unmarshal(body, &data)

		if data["name"] != requestData["name"] {
			t.Error("Request body name mismatch")
		}

		// Send response
		w.Header().Set("Content-Type", ContentTypeJSON)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(responseData)
	}))
	defer server.Close()

	client := NewClient()
	var result map[string]interface{}

	err := client.PostJSON(server.URL, requestData, &result)
	if err != nil {
		t.Fatalf("PostJSON() failed: %v", err)
	}

	if result["id"] != responseData["id"] {
		t.Errorf("Expected id=%v, got %v", responseData["id"], result["id"])
	}
	if result["name"] != responseData["name"] {
		t.Errorf("Expected name=%s, got %s", responseData["name"], result["name"])
	}
}

func TestClient_PostJSON_NilResult(t *testing.T) {
	requestData := map[string]string{
		"action": "delete",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient()

	// Pass nil for result when we don't care about the response
	err := client.PostJSON(server.URL, requestData, nil)
	if err != nil {
		t.Fatalf("PostJSON() with nil result failed: %v", err)
	}
}

func TestClient_PostJSON_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "value" {
			t.Error("Expected X-Custom header")
		}
		w.Header().Set("Content-Type", ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewClient()
	var result map[string]string

	err := client.PostJSON(
		server.URL,
		map[string]string{"data": "test"},
		&result,
		Header("X-Custom", "value"),
	)

	if err != nil {
		t.Fatalf("PostJSON() with options failed: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("Expected status=ok, got %s", result["status"])
	}
}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func TestClient_GetJSON_WithStruct(t *testing.T) {
	expectedUser := User{
		ID:    456,
		Name:  "Charlie",
		Email: "charlie@example.com",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", ContentTypeJSON)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(expectedUser)
	}))
	defer server.Close()

	client := NewClient()
	var result User

	err := client.GetJSON(server.URL, &result)
	if err != nil {
		t.Fatalf("GetJSON() with struct failed: %v", err)
	}

	if result.ID != expectedUser.ID {
		t.Errorf("Expected ID=%d, got %d", expectedUser.ID, result.ID)
	}
	if result.Name != expectedUser.Name {
		t.Errorf("Expected Name=%s, got %s", expectedUser.Name, result.Name)
	}
	if result.Email != expectedUser.Email {
		t.Errorf("Expected Email=%s, got %s", expectedUser.Email, result.Email)
	}
}

func TestClient_PostJSON_WithStruct(t *testing.T) {
	requestUser := User{
		Name:  "David",
		Email: "david@example.com",
	}

	responseUser := User{
		ID:    789,
		Name:  "David",
		Email: "david@example.com",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var receivedUser User
		json.NewDecoder(r.Body).Decode(&receivedUser)

		if receivedUser.Name != requestUser.Name {
			t.Error("Request body name mismatch")
		}

		w.Header().Set("Content-Type", ContentTypeJSON)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(responseUser)
	}))
	defer server.Close()

	client := NewClient()
	var result User

	err := client.PostJSON(server.URL, requestUser, &result)
	if err != nil {
		t.Fatalf("PostJSON() with struct failed: %v", err)
	}

	if result.ID != responseUser.ID {
		t.Errorf("Expected ID=%d, got %d", responseUser.ID, result.ID)
	}
}

func TestRequestBuilder_JSON_Chaining(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != ContentTypeJSON {
			t.Error("Expected Content-Type to be application/json")
		}

		body, _ := io.ReadAll(r.Body)
		var data map[string]string
		json.Unmarshal(body, &data)

		if data["key"] != "value" {
			t.Error("Expected key=value in body")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	resp, err := client.NewRequest().
		Method("POST").
		URL(server.URL).
		JSON(map[string]string{"key": "value"}).
		Do()

	if err != nil {
		t.Fatalf("Request with JSON() failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestClient_GetJSON_NetworkError(t *testing.T) {
	client := NewClient()
	var result map[string]interface{}

	// Invalid URL will cause network error
	err := client.GetJSON("http://invalid-domain-that-does-not-exist-12345.com/data", &result)
	if err == nil {
		t.Error("Expected network error for invalid domain")
	}
}

func TestClient_PostJSON_NetworkError(t *testing.T) {
	client := NewClient()
	var result map[string]interface{}

	err := client.PostJSON(
		"http://invalid-domain-that-does-not-exist-12345.com/data",
		map[string]string{"test": "data"},
		&result,
	)
	if err == nil {
		t.Error("Expected network error for invalid domain")
	}
}
