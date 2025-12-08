package httpc

import (
	"bytes"
	"compress/gzip"
	"testing"
)

func TestDecodeGzipBody(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "Valid gzip data",
			input:       "Hello, World!",
			expectError: false,
		},
		{
			name:        "Empty data",
			input:       "",
			expectError: false,
		},
		{
			name:        "Large data",
			input:       string(bytes.Repeat([]byte("Test data "), 1000)),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compress the data
			var buf bytes.Buffer
			if len(tt.input) > 0 {
				gzWriter := gzip.NewWriter(&buf)
				_, err := gzWriter.Write([]byte(tt.input))
				if err != nil {
					t.Fatalf("Failed to compress data: %v", err)
				}
				gzWriter.Close()
			}

			// Decode the data
			decoded, err := decodeGzipBody(buf.Bytes())

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError && string(decoded) != tt.input {
				t.Errorf("Decoded data mismatch. Got %q, want %q", string(decoded), tt.input)
			}
		})
	}
}

func TestDecodeGzipBody_InvalidData(t *testing.T) {
	// Test with invalid gzip data
	invalidData := []byte("this is not gzip data")
	decoded, err := decodeGzipBody(invalidData)

	// Should return the original data on error
	if err == nil {
		t.Error("Expected error for invalid gzip data")
	}

	if !bytes.Equal(decoded, invalidData) {
		t.Error("Should return original data when decompression fails")
	}
}

func TestIsGzipEncoded(t *testing.T) {
	tests := []struct {
		encoding string
		expected bool
	}{
		{"gzip", true},
		{"GZIP", true},
		{"Gzip", true},
		{"GzIp", true},
		{"deflate", false},
		{"br", false},
		{"", false},
		{"gzip, deflate", false}, // Not exactly "gzip"
	}

	for _, tt := range tests {
		t.Run(tt.encoding, func(t *testing.T) {
			got := isGzipEncoded(tt.encoding)
			if got != tt.expected {
				t.Errorf("isGzipEncoded(%q) = %v, want %v", tt.encoding, got, tt.expected)
			}
		})
	}
}
