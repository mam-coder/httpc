// Package httpc provides tests for HTTP response handling.
// This file contains tests for response body reading (Bytes, String, JSON),
// caching behavior, error handling, and success status checking.
package httpc

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestResponse_Bytes(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name:    "Read simple string body",
			body:    "hello world",
			wantErr: false,
		},
		{
			name:    "Read empty body",
			body:    "",
			wantErr: false,
		},
		{
			name:    "Read JSON body",
			body:    `{"name":"John","age":30}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &Response{
				Response: &http.Response{
					Body: io.NopCloser(strings.NewReader(tt.body)),
				},
			}

			got, err := resp.Bytes()
			if (err != nil) != tt.wantErr {
				t.Errorf("Bytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && string(got) != tt.body {
				t.Errorf("Bytes() = %v, want %v", string(got), tt.body)
			}
		})
	}
}

func TestResponse_Bytes_CachesResult(t *testing.T) {
	body := "test body"
	resp := &Response{
		Response: &http.Response{
			Body: io.NopCloser(strings.NewReader(body)),
		},
	}

	// First call
	got1, err := resp.Bytes()
	if err != nil {
		t.Fatalf("First Bytes() call failed: %v", err)
	}

	// Second call should return cached result
	got2, err := resp.Bytes()
	if err != nil {
		t.Fatalf("Second Bytes() call failed: %v", err)
	}

	if string(got1) != string(got2) {
		t.Errorf("Cached result differs: first=%s, second=%s", got1, got2)
	}

	if string(got1) != body {
		t.Errorf("Bytes() = %s, want %s", got1, body)
	}
}

func TestResponse_String(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name:    "Convert simple string",
			body:    "hello world",
			wantErr: false,
		},
		{
			name:    "Convert empty string",
			body:    "",
			wantErr: false,
		},
		{
			name:    "Convert multiline string",
			body:    "line1\nline2\nline3",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &Response{
				Response: &http.Response{
					Body: io.NopCloser(strings.NewReader(tt.body)),
				},
			}

			got, err := resp.String()
			if (err != nil) != tt.wantErr {
				t.Errorf("String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.body {
				t.Errorf("String() = %v, want %v", got, tt.body)
			}
		})
	}
}

func TestResponse_JSON(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		target  interface{}
		wantErr bool
		check   func(t *testing.T, v interface{})
	}{
		{
			name:    "Unmarshal simple object",
			body:    `{"name":"John","age":30}`,
			target:  &map[string]interface{}{},
			wantErr: false,
			check: func(t *testing.T, v interface{}) {
				m := v.(*map[string]interface{})
				if (*m)["name"] != "John" {
					t.Errorf("Expected name=John, got %v", (*m)["name"])
				}
				if (*m)["age"].(float64) != 30 {
					t.Errorf("Expected age=30, got %v", (*m)["age"])
				}
			},
		},
		{
			name:    "Unmarshal array",
			body:    `[1,2,3]`,
			target:  &[]int{},
			wantErr: false,
			check: func(t *testing.T, v interface{}) {
				arr := v.(*[]int)
				if len(*arr) != 3 {
					t.Errorf("Expected length 3, got %d", len(*arr))
				}
			},
		},
		{
			name:    "Invalid JSON",
			body:    `{invalid json}`,
			target:  &map[string]interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &Response{
				Response: &http.Response{
					Body: io.NopCloser(strings.NewReader(tt.body)),
				},
			}

			err := resp.JSON(tt.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, tt.target)
			}
		})
	}
}

func TestResponse_isSuccess(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"200 OK", 200, true},
		{"201 Created", 201, true},
		{"204 No Content", 204, true},
		{"299 Edge of 2xx", 299, true},
		{"300 Multiple Choices", 300, false},
		{"400 Bad Request", 400, false},
		{"404 Not Found", 404, false},
		{"500 Internal Server Error", 500, false},
		{"199 Before 2xx", 199, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &Response{
				Response: &http.Response{
					StatusCode: tt.statusCode,
				},
			}

			if got := resp.isSuccess(); got != tt.want {
				t.Errorf("isSuccess() = %v, want %v for status code %d", got, tt.want, tt.statusCode)
			}
		})
	}
}

type testStruct struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func TestResponse_JSON_WithStruct(t *testing.T) {
	body := `{"name":"Alice","email":"alice@example.com","age":25}`
	resp := &Response{
		Response: &http.Response{
			Body: io.NopCloser(strings.NewReader(body)),
		},
	}

	var result testStruct
	err := resp.JSON(&result)
	if err != nil {
		t.Fatalf("JSON() failed: %v", err)
	}

	if result.Name != "Alice" {
		t.Errorf("Expected name=Alice, got %s", result.Name)
	}
	if result.Email != "alice@example.com" {
		t.Errorf("Expected email=alice@example.com, got %s", result.Email)
	}
	if result.Age != 25 {
		t.Errorf("Expected age=25, got %d", result.Age)
	}
}

func TestResponse_String_FromCache(t *testing.T) {
	body := "test body"
	resp := &Response{
		Response: &http.Response{
			Body: io.NopCloser(strings.NewReader(body)),
		},
	}

	// First call reads from body
	str1, err := resp.String()
	if err != nil {
		t.Fatalf("First String() call failed: %v", err)
	}

	// Second call should use cache
	str2, err := resp.String()
	if err != nil {
		t.Fatalf("Second String() call failed: %v", err)
	}

	if str1 != str2 {
		t.Error("Cached string differs from original")
	}
}

func TestResponse_JSON_FromCache(t *testing.T) {
	body := `{"name":"Bob"}`
	resp := &Response{
		Response: &http.Response{
			Body: io.NopCloser(strings.NewReader(body)),
		},
	}

	var result1 map[string]string
	err := resp.JSON(&result1)
	if err != nil {
		t.Fatalf("First JSON() call failed: %v", err)
	}

	// Second call should use cached body
	var result2 map[string]string
	err = resp.JSON(&result2)
	if err != nil {
		t.Fatalf("Second JSON() call failed: %v", err)
	}

	if result1["name"] != result2["name"] {
		t.Error("Cached JSON differs from original")
	}
}

// errorReader always returns an error when read
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func (e *errorReader) Close() error {
	return nil
}

func TestResponse_Bytes_ReadError(t *testing.T) {
	resp := &Response{
		Response: &http.Response{
			Body: &errorReader{},
		},
	}

	_, err := resp.Bytes()
	if err == nil {
		t.Error("Expected error when reading from error reader")
	}
}

func TestResponse_String_ReadError(t *testing.T) {
	resp := &Response{
		Response: &http.Response{
			Body: &errorReader{},
		},
	}

	_, err := resp.String()
	if err == nil {
		t.Error("Expected error when reading from error reader")
	}
}

func TestResponse_JSON_ReadError(t *testing.T) {
	resp := &Response{
		Response: &http.Response{
			Body: &errorReader{},
		},
	}

	var result map[string]interface{}
	err := resp.JSON(&result)
	if err == nil {
		t.Error("Expected error when reading from error reader")
	}
}
