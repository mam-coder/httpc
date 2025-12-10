package httpc

import (
	"context"
	"errors"
	"fmt"
	"net"
)

// Error represents an HTTP error response from the client.
// It contains the HTTP status code, error message, and response body.
//
// Example:
//
//	resp, err := client.Get("/api/users")
//	if err != nil {
//		var httpErr *httpc.Error
//		if errors.As(err, &httpErr) {
//			fmt.Printf("Status: %d, Message: %s\n", httpErr.StatusCode, httpErr.Message)
//		}
//	}
type Error struct {
	// StatusCode is the HTTP status code returned by the server
	StatusCode int

	// Message is a human-readable error message
	Message string

	// Body is the raw response body from the server
	Body []byte
}

// Error implements the error interface and returns a formatted error message
// containing the status code and message.
func (e *Error) Error() string {
	return fmt.Sprintf("Request failed with status %d: %s", e.StatusCode, e.Message)
}

// IsTimeout checks if the error is a timeout error.
// It returns true if the error is either a net.Error with Timeout() == true
// or a context.DeadlineExceeded error.
//
// Example:
//
//	resp, err := client.Get("/api/slow-endpoint")
//	if httpc.IsTimeout(err) {
//		log.Println("Request timed out")
//	}
func IsTimeout(err error) bool {
	if err == nil {
		return false
	}

	// Check for the net.Error with Timeout() method
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	// Check for context deadline exceeded
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return false
}
