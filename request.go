package httpc

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RequestBuilder provides a fluent interface for building HTTP requests.
// It allows chaining method calls to configure all aspects of an HTTP request
// before executing it with Do().
//
// Example:
//
//	resp, err := client.NewRequest().
//	    Method("POST").
//	    URL("/api/users").
//	    Header("Authorization", "Bearer token").
//	    Query("include", "profile").
//	    JSON(userData).
//	    Timeout(10*time.Second).
//	    Do()
type RequestBuilder struct {
	client  *Client
	method  string
	url     string
	headers map[string]string
	query   url.Values
	body    io.Reader
	timeout time.Duration
	ctx     context.Context
	err     error
}

// NewRequest creates a new RequestBuilder for building and executing HTTP requests.
//
// Example:
//
//	rb := client.NewRequest()
//	resp, err := rb.Method("GET").URL("/api/users").Do()
func (c *Client) NewRequest() *RequestBuilder {
	return &RequestBuilder{
		client:  c,
		headers: make(map[string]string),
		query:   make(url.Values),
		ctx:     context.Background(),
	}
}

// Method sets the HTTP method for the request (GET, POST, PUT, DELETE, PATCH, etc.).
//
// Example:
//
//	rb.Method("POST")
func (rb *RequestBuilder) Method(method string) *RequestBuilder {
	rb.method = method
	return rb
}

// URL sets the request URL. Can be absolute or relative to the client's base URL.
//
// Example:
//
//	rb.URL("/api/users") // Relative to base URL
//	rb.URL("https://api.example.com/users") // Absolute URL
func (rb *RequestBuilder) URL(url string) *RequestBuilder {
	rb.url = url
	return rb
}

// Header adds a header to the request. Can be called multiple times to add multiple headers.
//
// Example:
//
//	rb.Header("Authorization", "Bearer token").
//	   Header("Accept", httpc.ContentTypeJSON)
func (rb *RequestBuilder) Header(key, value string) *RequestBuilder {
	rb.headers[key] = value
	return rb
}

// Query adds a query parameter to the request. Can be called multiple times.
// Values with the same key will be appended (e.g., ?tag=a&tag=b).
//
// Example:
//
//	rb.Query("page", "1").Query("limit", "10")
func (rb *RequestBuilder) Query(key, value string) *RequestBuilder {
	if rb.query == nil {
		rb.query = make(url.Values)
	}
	rb.query.Add(key, value)
	return rb
}

// QueryParams adds multiple query parameters from a map. Each key-value pair
// will be set (replacing any existing values for those keys).
//
// Example:
//
//	rb.QueryParams(map[string]string{
//	    "page": "1",
//	    "limit": "10",
//	    "sort": "name",
//	})
func (rb *RequestBuilder) QueryParams(params map[string]string) *RequestBuilder {
	if rb.query == nil {
		rb.query = make(url.Values)
	}
	for key, value := range params {
		rb.query.Set(key, value)
	}
	return rb
}

// Body sets the request body from an io.Reader.
// For JSON bodies, use the JSON() method instead.
//
// Example:
//
//	rb.Body(strings.NewReader("plain text body"))
func (rb *RequestBuilder) Body(body io.Reader) *RequestBuilder {
	rb.body = body
	return rb
}

// Timeout sets a request-specific timeout, overriding the client's default timeout.
// This creates a timeout context that will cancel the request if it exceeds the duration.
// The timeout applies to the entire request/response cycle, including connection time,
// redirects, and reading the response body.
//
// Example:
//
//	rb.Timeout(5 * time.Second)
func (rb *RequestBuilder) Timeout(timeout time.Duration) *RequestBuilder {
	rb.timeout = timeout
	return rb
}

// Context sets the context for the request, allowing for cancellation and deadlines.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//	rb.Context(ctx)
func (rb *RequestBuilder) Context(ctx context.Context) *RequestBuilder {
	rb.ctx = ctx
	return rb
}

// buildURL builds the request URL
func (rb *RequestBuilder) buildURL() string {
	fullURL := rb.resolveURL()

	if len(rb.query) > 0 {
		parsedUrl, err := url.Parse(fullURL)

		// If there are no query parameters, return the original URL
		if err != nil {
			return fullURL
		}

		// Get the query parameters
		currentQuery := parsedUrl.Query()

		for key, values := range rb.query {
			for _, value := range values {
				currentQuery.Add(key, value)
			}
		}

		parsedUrl.RawQuery = currentQuery.Encode()

		return parsedUrl.String()
	}

	return fullURL
}

// resolveURL resolves the request URL
func (rb *RequestBuilder) resolveURL() string {
	baseURL := strings.TrimSpace(rb.client.baseURL)
	requestURL := strings.TrimSpace(rb.url)

	// If URL is absolute, return it
	if isAbsoluteURL(requestURL) {
		return requestURL
	}
	// If the baseURL is empty, return the requestURL
	if baseURL == "" {
		return requestURL
	}

	// Combine the baseURL and requestURL

	baseURL = strings.TrimSuffix(baseURL, "/")

	if requestURL == "" {
		return baseURL
	}

	if !strings.HasPrefix(requestURL, "/") {
		return baseURL + "/" + requestURL
	}

	return baseURL + requestURL

}

// isAbsoluteURL checks if URL is absolute
func isAbsoluteURL(urlStr string) bool {
	return strings.HasPrefix(urlStr, "http://") ||
		strings.HasPrefix(urlStr, "https://")
}

// applyHeaders applies headers to the request
func (rb *RequestBuilder) applyHeaders(req *http.Request) {
	// Apply client Default headers
	rb.client.mu.Lock()
	//defer rb.client.mu.Unlock()
	for key, value := range rb.client.headers {
		req.Header.Set(key, value)
	}
	rb.client.mu.Unlock()

	// Apply request specific headers
	for key, value := range rb.headers {
		req.Header.Set(key, value)
	}

	rb.setDefaultHeaders(req)
}

// setDefaultHeaders sets default headers if they are not set
func (rb *RequestBuilder) setDefaultHeaders(req *http.Request) {
	// User-Agent
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", "go-httpc/1.0")
	}
	// Accept
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "*/*")
	}
	//
	//Accept-Encoding
	if req.Header.Get("Accept-Encoding") == "" {
		req.Header.Set("Accept-Encoding", "*")
	}
	//
	//Content-Type
	if req.Body != nil && req.Header.Get("Content-Type") == "" {
		// Don't set the default Content-Type, let it be set explicitly
		// or by JSON() method
	}
}

// Do executes the HTTP request and returns the response.
// This should be called as the final method in the RequestBuilder chain.
// Any errors that occurred during request building will be returned here.
//
// Example:
//
//	resp, err := client.NewRequest().
//	    Method("GET").
//	    URL("/api/users").
//	    Do()
func (rb *RequestBuilder) Do() (*Response, error) {
	if rb.err != nil {
		return nil, rb.err
	}

	//build the full URL
	fullURL := rb.buildURL()

	// Apply timeout to context if specified
	ctx := rb.ctx
	var cancel context.CancelFunc
	if rb.timeout > 0 {
		ctx, cancel = context.WithTimeout(rb.ctx, rb.timeout)
		defer cancel()
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, rb.method, fullURL, rb.body)

	if err != nil {
		return nil, err
	}

	//Apply headers
	rb.applyHeaders(req)

	return rb.client.doRequest(req)

}
