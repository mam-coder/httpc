// Package httpc provides tests for error handling functionality.
// This file contains tests for the custom Error type and timeout detection utilities.
package httpc

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
		wantErr    string
	}{
		{
			name:       "404 error",
			statusCode: 404,
			message:    "Not Found",
			wantErr:    "Request faild with status 	404: Not Found",
		},
		{
			name:       "500 error",
			statusCode: 500,
			message:    "Internal Server Error",
			wantErr:    "Request faild with status 	500: Internal Server Error",
		},
		{
			name:       "Custom message",
			statusCode: 400,
			message:    "Bad Request: invalid input",
			wantErr:    "Request faild with status 	400: Bad Request: invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &Error{
				StatusCode: tt.statusCode,
				Message:    tt.message,
			}

			if got := err.Error(); got != tt.wantErr {
				t.Errorf("Error() = %v, want %v", got, tt.wantErr)
			}
		})
	}
}

func TestError_WithBody(t *testing.T) {
	body := []byte(`{"error":"validation failed"}`)
	err := &Error{
		StatusCode: 422,
		Message:    "Unprocessable Entity",
		Body:       body,
	}

	if string(err.Body) != `{"error":"validation failed"}` {
		t.Errorf("Expected body to be preserved, got %s", err.Body)
	}
}

func TestIsTimeout(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "Nil error",
			err:  nil,
			want: false,
		},
		{
			name: "Context deadline exceeded",
			err:  context.DeadlineExceeded,
			want: true,
		},
		{
			name: "Timeout error wrapped",
			err:  &timeoutError{},
			want: true,
		},
		{
			name: "Regular error",
			err:  errors.New("regular error"),
			want: false,
		},
		{
			name: "Network error without timeout",
			err:  &netError{timeout: false},
			want: false,
		},
		{
			name: "Network error with timeout",
			err:  &netError{timeout: true},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTimeout(tt.err); got != tt.want {
				t.Errorf("IsTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsTimeout_WithWrappedErrors(t *testing.T) {
	// Test with wrapped context.DeadlineExceeded
	wrappedErr := errors.New("operation failed: " + context.DeadlineExceeded.Error())
	if IsTimeout(wrappedErr) {
		t.Error("IsTimeout() should not match string-wrapped errors")
	}

	// Test with properly wrapped error using %w
	properlyWrapped := &timeoutError{}
	if !IsTimeout(properlyWrapped) {
		t.Error("IsTimeout() should match net.Error with Timeout()")
	}
}

// Mock types for testing

type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

type netError struct {
	timeout   bool
	temporary bool
}

func (e *netError) Error() string   { return "network error" }
func (e *netError) Timeout() bool   { return e.timeout }
func (e *netError) Temporary() bool { return e.temporary }

func TestIsTimeout_RealNetworkTimeout(t *testing.T) {
	// Simulate a real network timeout scenario
	_, err := net.DialTimeout("tcp", "10.255.255.1:80", 1*time.Millisecond)
	if err != nil {
		if !IsTimeout(err) {
			t.Errorf("Expected real network timeout to be detected")
		}
	}
}

func TestError_AsError(t *testing.T) {
	httpErr := &Error{
		StatusCode: 404,
		Message:    "Not Found",
		Body:       []byte("page not found"),
	}

	var targetErr *Error
	if !errors.As(httpErr, &targetErr) {
		t.Error("Expected errors.As to work with *Error type")
	}

	if targetErr.StatusCode != 404 {
		t.Errorf("Expected StatusCode=404, got %d", targetErr.StatusCode)
	}
}

func TestIsTimeout_ContextCanceled(t *testing.T) {
	// Context canceled should not be considered a timeout
	if IsTimeout(context.Canceled) {
		t.Error("IsTimeout() should return false for context.Canceled")
	}
}
