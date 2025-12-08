// Package httpc provides a modern, fluent HTTP client for Go with support for
// retry logic, interceptors, and convenient request building.
//
// # Features
//
//   - Fluent API with chainable methods for building requests
//   - Automatic retry with configurable exponential backoff
//   - Interceptor pattern for request modification (auth, logging, rate limiting, etc.)
//   - Built-in JSON marshaling and unmarshaling
//   - Full context.Context support with cancellation, timeouts, and deadlines
//   - All HTTP methods (GET, POST, PUT, DELETE, PATCH) with context variants
//   - Thread-safe client that can be used concurrently
//   - HTTP/2 support with automatic fallback to HTTP/1.1
//   - Connection pooling and keep-alive
//
// # Quick Start
//
// Create a client and make a simple GET request:
//
//	client := httpc.NewClient()
//	resp, err := client.Get("https://api.example.com/users")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	body, _ := resp.String()
//	fmt.Println(body)
//
// # Client Configuration
//
// Configure a client with base URL, timeout, and headers:
//
//	retryConfig := httpc.DefaultRetryConfig()
//	retryConfig.MaxRetries = 3
//	retryConfig.Backoff = time.Second
//
//	client := httpc.NewClient(
//	    httpc.WithBaseURL("https://api.example.com"),
//	    httpc.WithTimeout(30*time.Second),
//	    httpc.WithHeader("User-Agent", "MyApp/1.0"),
//	    httpc.WithRetry(*retryConfig),
//	)
//
// # JSON Requests
//
// Send and receive JSON easily:
//
//	type User struct {
//	    Name  string `json:"name"`
//	    Email string `json:"email"`
//	}
//
//	user := User{Name: "John", Email: "john@example.com"}
//	var result User
//	err := client.PostJSON("/users", user, &result)
//
// # Request Builder
//
// Use the fluent request builder for complex requests:
//
//	resp, err := client.NewRequest().
//	    Method("POST").
//	    URL("/api/users").
//	    Header("Authorization", "Bearer token").
//	    Query("include", "profile").
//	    JSON(userData).
//	    Timeout(10*time.Second).
//	    Do()
//
// # Interceptors
//
// Add interceptors for cross-cutting concerns:
//
//	client := httpc.NewClient(
//	    httpc.WithDebug(),
//	    httpc.WithAuthorization("your-token"),
//	    httpc.WithLogger(log.Default()),
//	)
//
// Available built-in options:
//   - WithAuthorization(token): Bearer token authentication
//   - WithBaseAuth(user, pass): HTTP Basic authentication
//   - WithApiKey(header, key): API key authentication
//   - WithUserAgent(ua): Sets User-Agent header
//   - WithRequestId(header): Adds unique request IDs
//   - WithLogger(logger): Request/response logging
//   - WithDebug(): Debug mode with detailed logging
//   - WithBlockedList(domains): Blocks requests to specific domains
//   - WithHeaders(map): Adds custom headers
//
// Custom interceptors wrap the underlying http.RoundTripper:
//
//	customInterceptor := func(rt http.RoundTripper) http.RoundTripper {
//	    return &customTransport{transport: rt}
//	}
//	client := httpc.NewClient(httpc.WithInterceptor(customInterceptor))
//
// # Retry Logic
//
// Configure automatic retries with exponential backoff:
//
//	retryConfig := httpc.DefaultRetryConfig()
//	retryConfig.MaxRetries = 3
//	retryConfig.Backoff = time.Second
//
//	client := httpc.NewClient(
//	    httpc.WithRetry(*retryConfig),
//	)
//
// By default, requests are retried on:
//   - Network errors and timeouts
//   - 5xx server errors
//   - 429 (rate limit) responses
//
// # Error Handling
//
// Check for both network errors and HTTP errors:
//
//	resp, err := client.Get("/api/users")
//	if err != nil {
//	    if httpc.IsTimeout(err) {
//	        log.Println("Request timed out")
//	    }
//	    log.Fatal(err)
//	}
//
//	if resp.StatusCode >= 400 {
//	    log.Printf("HTTP error: %d", resp.StatusCode)
//	}
//
// # Context Support
//
// All HTTP methods support context for cancellation, timeouts, and deadlines:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	// All methods have WithContext variants
//	resp, err := client.GetWithContext(ctx, "/api/users")
//	resp, err := client.PostWithContext(ctx, "/api/users", userData)
//	resp, err := client.PutWithContext(ctx, "/api/users/123", userData)
//	resp, err := client.DeleteWithContext(ctx, "/api/users/123")
//	resp, err := client.PatchWithContext(ctx, "/api/users/123", updates)
//
//	// JSON methods also support context
//	var users []User
//	err := client.GetJSONWithContext(ctx, "/api/users", &users)
//
// Or use the request builder:
//
//	resp, err := client.NewRequest().
//	    Method("GET").
//	    URL("/api/users").
//	    Context(ctx).
//	    Timeout(5*time.Second).  // Timeout wraps the context
//	    Do()
//
// Or pass context as a RequestOption:
//
//	resp, err := client.Get("/api/users",
//	    httpc.WithContext(ctx),
//	    httpc.WithQuery("page", "1"),
//	)
//
// # Content Types
//
// The package provides constants for common content types:
//
//	httpc.ContentTypeJSON        // application/json
//	httpc.ContentTypeXML         // application/xml
//	httpc.ContentTypeForm        // application/x-www-form-urlencoded
//	httpc.ContentTypePlainText   // text/plain
//	httpc.ContentTypeHTML        // text/html
//	// ... and more
//
// # Thread Safety
//
// The Client is safe for concurrent use. You should create one client and
// reuse it for multiple requests to benefit from connection pooling:
//
//	// Create once, use many times from different goroutines
//	client := httpc.NewClient()
//
//	go func() {
//	    resp, _ := client.Get("/api/endpoint1")
//	    // ...
//	}()
//
//	go func() {
//	    resp, _ := client.Get("/api/endpoint2")
//	    // ...
//	}()
//
// # Best Practices
//
//   - Reuse clients to benefit from connection pooling
//   - Always use contexts for production code
//   - Set appropriate timeouts to prevent hanging requests
//   - Enable retry logic for transient failures
//   - Use interceptors to centralize cross-cutting concerns
//   - Check both error return values and HTTP status codes
package httpc
