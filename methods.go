package httpc

import "context"

// Get sends an HTTP GET request to the specified URL.
// Optional RequestOption functions can be provided to customize the request.
//
// Example:
//
//	resp, err := client.Get("/api/users",
//	    httpc.Header("Accept", httpc.ContentTypeJSON),
//	    httpc.WithQuery("page", "1"),
//	)
func (c *Client) Get(url string, opts ...RequestOption) (*Response, error) {
	rb := c.NewRequest().Method("GET").URL(url)
	for _, opt := range opts {
		opt(rb)
	}
	return rb.Do()
}

// Post sends an HTTP POST request with a JSON body to the specified URL.
// The body is automatically marshaled to JSON and the Content-Type header is set.
// Pass nil for body if no request body is needed.
//
// Example:
//
//	user := map[string]string{"name": "John", "email": "john@example.com"}
//	resp, err := client.Post("/api/users", user)
func (c *Client) Post(url string, body interface{}, opts ...RequestOption) (*Response, error) {
	rb := c.NewRequest().Method("POST").URL(url)

	if body != nil {
		rb.JSON(body)
	}

	for _, opt := range opts {
		opt(rb)
	}

	return rb.Do()
}

// Put sends an HTTP PUT request with a JSON body to the specified URL.
// The body is automatically marshaled to JSON and the Content-Type header is set.
// Pass nil for body if no request body is needed.
//
// Example:
//
//	user := map[string]string{"name": "John Doe"}
//	resp, err := client.Put("/api/users/123", user)
func (c *Client) Put(url string, body interface{}, opts ...RequestOption) (*Response, error) {
	rb := c.NewRequest().Method("PUT").URL(url)
	if body != nil {
		rb.JSON(body)
	}
	for _, opt := range opts {
		opt(rb)
	}
	return rb.Do()
}

// Delete sends an HTTP DELETE request to the specified URL.
//
// Example:
//
//	resp, err := client.Delete("/api/users/123")
func (c *Client) Delete(url string, opts ...RequestOption) (*Response, error) {
	rb := c.NewRequest().Method("DELETE").URL(url)
	for _, opt := range opts {
		opt(rb)
	}
	return rb.Do()
}

// Patch sends an HTTP PATCH request with a JSON body to the specified URL.
// The body is automatically marshaled to JSON and the Content-Type header is set.
// Pass nil for body if no request body is needed.
//
// Example:
//
//	updates := map[string]string{"status": "active"}
//	resp, err := client.Patch("/api/users/123", updates)
func (c *Client) Patch(url string, body interface{}, opts ...RequestOption) (*Response, error) {
	rb := c.NewRequest().Method("PATCH").URL(url)
	if body != nil {
		rb.JSON(body)
	}
	for _, opt := range opts {
		opt(rb)
	}
	return rb.Do()
}

// GetWithContext sends an HTTP GET request with a context.
// The context can be used for cancellation or setting deadlines.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	resp, err := client.GetWithContext(ctx, "/api/users")
func (c *Client) GetWithContext(ctx context.Context, url string, opts ...RequestOption) (*Response, error) {
	rb := c.NewRequest().Method("GET").URL(url).Context(ctx)
	for _, opt := range opts {
		opt(rb)
	}
	return rb.Do()
}

// PostWithContext sends an HTTP POST request with a context and JSON body.
// The context can be used for cancellation or setting deadlines.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	resp, err := client.PostWithContext(ctx, "/api/users", userData)
func (c *Client) PostWithContext(ctx context.Context, url string, body interface{}, opts ...RequestOption) (*Response, error) {
	rb := c.NewRequest().Method("POST").URL(url).Context(ctx)
	if body != nil {
		rb.JSON(body)
	}
	for _, opt := range opts {
		opt(rb)
	}
	return rb.Do()
}

// PutWithContext sends an HTTP PUT request with a context and JSON body.
// The context can be used for cancellation or setting deadlines.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	resp, err := client.PutWithContext(ctx, "/api/users/123", userData)
func (c *Client) PutWithContext(ctx context.Context, url string, body interface{}, opts ...RequestOption) (*Response, error) {
	rb := c.NewRequest().Method("PUT").URL(url).Context(ctx)
	if body != nil {
		rb.JSON(body)
	}
	for _, opt := range opts {
		opt(rb)
	}
	return rb.Do()
}

// DeleteWithContext sends an HTTP DELETE request with a context.
// The context can be used for cancellation or setting deadlines.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	resp, err := client.DeleteWithContext(ctx, "/api/users/123")
func (c *Client) DeleteWithContext(ctx context.Context, url string, opts ...RequestOption) (*Response, error) {
	rb := c.NewRequest().Method("DELETE").URL(url).Context(ctx)
	for _, opt := range opts {
		opt(rb)
	}
	return rb.Do()
}

// PatchWithContext sends an HTTP PATCH request with a context and JSON body.
// The context can be used for cancellation or setting deadlines.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	resp, err := client.PatchWithContext(ctx, "/api/users/123", updates)
func (c *Client) PatchWithContext(ctx context.Context, url string, body interface{}, opts ...RequestOption) (*Response, error) {
	rb := c.NewRequest().Method("PATCH").URL(url).Context(ctx)
	if body != nil {
		rb.JSON(body)
	}
	for _, opt := range opts {
		opt(rb)
	}
	return rb.Do()
}

// RequestOption is a function that configures a RequestBuilder.
// It allows customizing individual requests with headers, query parameters, etc.
type RequestOption func(*RequestBuilder)

// Header returns a RequestOption that adds a header to the request.
//
// Example:
//
//	resp, err := client.Get("/api/users",
//	    httpc.Header("Authorization", "Bearer token"),
//	    httpc.Header("Accept", httpc.ContentTypeJSON),
//	)
func Header(key, value string) RequestOption {
	return func(rb *RequestBuilder) {
		rb.Header(key, value)
	}
}

// WithQuery returns a RequestOption that adds a query parameter to the request.
//
// Example:
//
//	resp, err := client.Get("/api/users",
//	    httpc.WithQuery("page", "1"),
//	    httpc.WithQuery("limit", "10"),
//	)
func WithQuery(key, value string) RequestOption {
	return func(rb *RequestBuilder) {
		rb.Query(key, value)
	}
}

// WithContext returns a RequestOption that sets the context for the request.
// This is useful when you want to pass a context along with other request options.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	resp, err := client.Get("/api/users",
//	    httpc.WithContext(ctx),
//	    httpc.WithQuery("page", "1"),
//	)
func WithContext(ctx context.Context) RequestOption {
	return func(rb *RequestBuilder) {
		rb.Context(ctx)
	}
}
