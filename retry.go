package httpc

import (
	"net/http"
	"time"
)

// RetryConfig configures the retry behavior for failed HTTP requests.
// It defines the maximum number of retries, backoff duration, and
// a custom condition function to determine if a request should be retried.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int

	// Backoff is the base duration between retries (multiplied by attempt number for exponential backoff)
	Backoff time.Duration

	// RetryIf is a function that determines whether a request should be retried
	// based on the response or error. If nil, defaultRetryCondition is used.
	RetryIf func(*http.Response, error) bool
}

// DefaultRetryConfig returns a RetryConfig with sensible defaults:
// - 3 maximum retries
// - 1 second base backoff
// - Retries on network errors, 5xx status codes, and 429 (rate limit)
//
// Example:
//
//	config := httpc.DefaultRetryConfig()
//	// Customize if needed
//	config.MaxRetries = 5
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries: 3,
		Backoff:    time.Second,
		RetryIf:    defaultRetryCondition,
	}
}

// defaultRetryCondition returns true if the request should be retried.
// It retries on any error, 5xx server errors, or 429 (rate limit) responses.
func defaultRetryCondition(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}

	// Retry on 5xx errors and 429 (rate limit)
	return resp.StatusCode >= 500 || resp.StatusCode == 429
}

// doRequest executes a single HTTP request without retry logic.
// It wraps the standard http.Client.Do method and returns a Response.
func (c *Client) doRequest(req *http.Request) (*Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return &Response{Response: resp}, nil
}
