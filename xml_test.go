// Package httpc provides tests for XML handling functionality.
// This file contains tests for XML marshaling/unmarshaling, convenience methods
// (GetXML, PostXML), and error handling for invalid XML.
package httpc

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type Config struct {
	XMLName     xml.Name `xml:"config"`
	Environment string   `xml:"environment"`
	Port        int      `xml:"port"`
}

type Settings struct {
	XMLName xml.Name `xml:"settings"`
	ID      int      `xml:"id"`
	Name    string   `xml:"name"`
	Enabled bool     `xml:"enabled"`
}

func TestRequestBuilder_XML(t *testing.T) {
	client := NewClient()

	config := Config{
		Environment: "production",
		Port:        8080,
	}

	rb := client.NewRequest().
		Method("POST").
		URL("http://example.com").
		XML(config)

	// Check that Content-Type header is set
	if rb.headers["Content-Type"] != ContentTypeXML {
		t.Errorf("Expected Content-Type %s, got %s", ContentTypeXML, rb.headers["Content-Type"])
	}

	// Check that body is set
	if rb.body == nil {
		t.Error("Expected body to be set")
	}
}

func TestRequestBuilder_XML_WithError(t *testing.T) {
	client := NewClient()

	// channels cannot be marshaled to XML
	invalidData := make(chan int)

	rb := client.NewRequest().
		Method("POST").
		URL("http://example.com").
		XML(invalidData)

	// Check that error is set
	if rb.err == nil {
		t.Error("Expected error when marshaling invalid data")
	}
}

func TestClient_GetXML(t *testing.T) {
	expectedConfig := Config{
		Environment: "development",
		Port:        3000,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}
		w.Header().Set("Content-Type", ContentTypeXML)
		w.WriteHeader(http.StatusOK)
		xml.NewEncoder(w).Encode(expectedConfig)
	}))
	defer server.Close()

	client := NewClient()
	var result Config

	err := client.GetXML(server.URL, &result)
	if err != nil {
		t.Fatalf("GetXML() failed: %v", err)
	}

	if result.Environment != expectedConfig.Environment {
		t.Errorf("Expected environment=%s, got %s", expectedConfig.Environment, result.Environment)
	}
	if result.Port != expectedConfig.Port {
		t.Errorf("Expected port=%d, got %d", expectedConfig.Port, result.Port)
	}
}

func TestClient_GetXML_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("env") != "production" {
			t.Error("Expected env=production query param")
		}
		w.Header().Set("Content-Type", ContentTypeXML)
		w.WriteHeader(http.StatusOK)
		config := Config{Environment: "production", Port: 8080}
		xml.NewEncoder(w).Encode(config)
	}))
	defer server.Close()

	client := NewClient()
	var result Config

	err := client.GetXML(server.URL, &result, WithQuery("env", "production"))
	if err != nil {
		t.Fatalf("GetXML() with options failed: %v", err)
	}

	if result.Environment != "production" {
		t.Errorf("Expected environment=production, got %s", result.Environment)
	}
}

func TestClient_GetXML_InvalidXML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid xml"))
	}))
	defer server.Close()

	client := NewClient()
	var result Config

	err := client.GetXML(server.URL, &result)
	if err == nil {
		t.Error("Expected error when parsing invalid XML")
	}
}

func TestClient_PostXML(t *testing.T) {
	requestConfig := Config{
		Environment: "staging",
		Port:        9090,
	}

	responseConfig := Config{
		Environment: "staging",
		Port:        9090,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Check Content-Type
		if r.Header.Get("Content-Type") != ContentTypeXML {
			t.Errorf("Expected Content-Type %s", ContentTypeXML)
		}

		// Read and verify request body
		body, _ := io.ReadAll(r.Body)
		var data Config
		xml.Unmarshal(body, &data)

		if data.Environment != requestConfig.Environment {
			t.Error("Request body environment mismatch")
		}

		// Send response
		w.Header().Set("Content-Type", ContentTypeXML)
		w.WriteHeader(http.StatusCreated)
		xml.NewEncoder(w).Encode(responseConfig)
	}))
	defer server.Close()

	client := NewClient()
	var result Config

	err := client.PostXML(server.URL, requestConfig, &result)
	if err != nil {
		t.Fatalf("PostXML() failed: %v", err)
	}

	if result.Environment != responseConfig.Environment {
		t.Errorf("Expected environment=%s, got %s", responseConfig.Environment, result.Environment)
	}
	if result.Port != responseConfig.Port {
		t.Errorf("Expected port=%d, got %d", responseConfig.Port, result.Port)
	}
}

func TestClient_PostXML_NilResult(t *testing.T) {
	requestConfig := Config{
		Environment: "test",
		Port:        5000,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient()

	// Pass nil for result when we don't care about the response
	err := client.PostXML(server.URL, requestConfig, nil)
	if err != nil {
		t.Fatalf("PostXML() with nil result failed: %v", err)
	}
}

func TestClient_PostXML_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "value" {
			t.Error("Expected X-Custom header")
		}
		w.Header().Set("Content-Type", ContentTypeXML)
		w.WriteHeader(http.StatusOK)
		config := Config{Environment: "production", Port: 8080}
		xml.NewEncoder(w).Encode(config)
	}))
	defer server.Close()

	client := NewClient()
	var result Config

	err := client.PostXML(
		server.URL,
		Config{Environment: "test", Port: 3000},
		&result,
		Header("X-Custom", "value"),
	)

	if err != nil {
		t.Fatalf("PostXML() with options failed: %v", err)
	}

	if result.Environment != "production" {
		t.Errorf("Expected environment=production, got %s", result.Environment)
	}
}

func TestClient_GetXML_WithStruct(t *testing.T) {
	expectedSettings := Settings{
		ID:      123,
		Name:    "TestSetting",
		Enabled: true,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", ContentTypeXML)
		w.WriteHeader(http.StatusOK)
		xml.NewEncoder(w).Encode(expectedSettings)
	}))
	defer server.Close()

	client := NewClient()
	var result Settings

	err := client.GetXML(server.URL, &result)
	if err != nil {
		t.Fatalf("GetXML() with struct failed: %v", err)
	}

	if result.ID != expectedSettings.ID {
		t.Errorf("Expected ID=%d, got %d", expectedSettings.ID, result.ID)
	}
	if result.Name != expectedSettings.Name {
		t.Errorf("Expected Name=%s, got %s", expectedSettings.Name, result.Name)
	}
	if result.Enabled != expectedSettings.Enabled {
		t.Errorf("Expected Enabled=%v, got %v", expectedSettings.Enabled, result.Enabled)
	}
}

func TestClient_PostXML_WithStruct(t *testing.T) {
	requestSettings := Settings{
		Name:    "NewSetting",
		Enabled: false,
	}

	responseSettings := Settings{
		ID:      456,
		Name:    "NewSetting",
		Enabled: false,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var receivedSettings Settings
		xml.NewDecoder(r.Body).Decode(&receivedSettings)

		if receivedSettings.Name != requestSettings.Name {
			t.Error("Request body name mismatch")
		}

		w.Header().Set("Content-Type", ContentTypeXML)
		w.WriteHeader(http.StatusCreated)
		xml.NewEncoder(w).Encode(responseSettings)
	}))
	defer server.Close()

	client := NewClient()
	var result Settings

	err := client.PostXML(server.URL, requestSettings, &result)
	if err != nil {
		t.Fatalf("PostXML() with struct failed: %v", err)
	}

	if result.ID != responseSettings.ID {
		t.Errorf("Expected ID=%d, got %d", responseSettings.ID, result.ID)
	}
}

func TestRequestBuilder_XML_Chaining(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != ContentTypeXML {
			t.Error("Expected Content-Type to be application/xml")
		}

		body, _ := io.ReadAll(r.Body)
		var data Config
		xml.Unmarshal(body, &data)

		if data.Environment != "production" {
			t.Error("Expected environment=production in body")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	resp, err := client.NewRequest().
		Method("POST").
		URL(server.URL).
		XML(Config{Environment: "production", Port: 8080}).
		Do()

	if err != nil {
		t.Fatalf("Request with XML() failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestClient_GetXML_NetworkError(t *testing.T) {
	client := NewClient()
	var result Config

	// Invalid URL will cause network error
	err := client.GetXML("http://invalid-domain-that-does-not-exist-12345.com/data", &result)
	if err == nil {
		t.Error("Expected network error for invalid domain")
	}
}

func TestClient_PostXML_NetworkError(t *testing.T) {
	client := NewClient()
	var result Config

	err := client.PostXML(
		"http://invalid-domain-that-does-not-exist-12345.com/data",
		Config{Environment: "test", Port: 3000},
		&result,
	)
	if err == nil {
		t.Error("Expected network error for invalid domain")
	}
}

func TestClient_GetXMLWithContext(t *testing.T) {
	expectedConfig := Config{
		Environment: "development",
		Port:        3000,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", ContentTypeXML)
		w.WriteHeader(http.StatusOK)
		xml.NewEncoder(w).Encode(expectedConfig)
	}))
	defer server.Close()

	client := NewClient()
	var result Config

	ctx := context.Background()
	err := client.GetXMLWithContext(ctx, server.URL, &result)
	if err != nil {
		t.Fatalf("GetXMLWithContext() failed: %v", err)
	}

	if result.Environment != expectedConfig.Environment {
		t.Errorf("Expected environment=%s, got %s", expectedConfig.Environment, result.Environment)
	}
}

func TestClient_GetXMLWithContext_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	var result Config

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := client.GetXMLWithContext(ctx, server.URL, &result)
	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestClient_PostXMLWithContext(t *testing.T) {
	requestConfig := Config{
		Environment: "staging",
		Port:        9090,
	}

	responseConfig := Config{
		Environment: "staging",
		Port:        9090,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", ContentTypeXML)
		w.WriteHeader(http.StatusCreated)
		xml.NewEncoder(w).Encode(responseConfig)
	}))
	defer server.Close()

	client := NewClient()
	var result Config

	ctx := context.Background()
	err := client.PostXMLWithContext(ctx, server.URL, requestConfig, &result)
	if err != nil {
		t.Fatalf("PostXMLWithContext() failed: %v", err)
	}

	if result.Environment != responseConfig.Environment {
		t.Errorf("Expected environment=%s, got %s", responseConfig.Environment, result.Environment)
	}
}

func TestClient_PostXMLWithContext_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient()
	var result Config

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := client.PostXMLWithContext(ctx, server.URL, Config{Environment: "test", Port: 3000}, &result)
	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestClient_PostXMLWithContext_NilResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient()

	ctx := context.Background()
	err := client.PostXMLWithContext(ctx, server.URL, Config{Environment: "test", Port: 3000}, nil)
	if err != nil {
		t.Fatalf("PostXMLWithContext() with nil result failed: %v", err)
	}
}

func TestClient_GetXMLWithContext_WithOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("env") != "production" {
			t.Error("Expected env=production query param")
		}
		w.Header().Set("Content-Type", ContentTypeXML)
		w.WriteHeader(http.StatusOK)
		config := Config{Environment: "production", Port: 8080}
		xml.NewEncoder(w).Encode(config)
	}))
	defer server.Close()

	client := NewClient()
	var result Config

	ctx := context.Background()
	err := client.GetXMLWithContext(ctx, server.URL, &result, WithQuery("env", "production"))
	if err != nil {
		t.Fatalf("GetXMLWithContext() with options failed: %v", err)
	}

	if result.Environment != "production" {
		t.Errorf("Expected environment=production, got %s", result.Environment)
	}
}
