# httpc

[![Go Reference](https://pkg.go.dev/badge/github.com/mam-coder/httpc.svg)](https://pkg.go.dev/github.com/mam-coder/httpc)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/mam-coder/httpc)](https://goreportcard.com/report/github.com/mam-coder/httpc)

A modern, fluent HTTP client library for Go with built-in retry logic, interceptors, and convenience methods.

## Features

- **Fluent API**: Chainable methods for building requests
- **Retry Logic**: Configurable automatic retry with exponential backoff
- **Interceptors**: Middleware pattern for request modification (auth, logging, rate limiting, etc.)
- **JSON Support**: Built-in JSON marshaling/unmarshaling
- **XML Support**: Built-in XML marshaling/unmarshaling
- **Context Support**: Full context.Context integration
- **Flexible Configuration**: Options pattern for client and request configuration
- **Thread-Safe**: Safe for concurrent use
- **HTTP/2 Support**: Automatic HTTP/2 with fallback to HTTP/1.1

## Installation

```bash
go get httpc
```

## Quick Start

### Simple GET Request

```go
client := httpc.NewClient()
resp, err := client.Get("https://api.example.com/users")
if err != nil {
    log.Fatal(err)
}

body, err := resp.String()
if err != nil {
    log.Fatal(err)
}
fmt.Println(body)
```

### JSON Request/Response

```go
client := httpc.NewClient(httpc.WithBaseURL("https://api.example.com"))

type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

// POST JSON
user := User{Name: "John", Email: "john@example.com"}
var result User
err := client.PostJSON("/users", user, &result)
if err != nil {
    log.Fatal(err)
}

// GET JSON
var users []User
err = client.GetJSON("/users", &users)
if err != nil {
    log.Fatal(err)
}
```

### XML Request/Response

```go
client := httpc.NewClient(httpc.WithBaseURL("https://api.example.com"))

type Config struct {
    Environment string `xml:"environment"`
    Port        int    `xml:"port"`
}

// POST XML
config := Config{Environment: "production", Port: 8080}
var result Config
err := client.PostXML("/config", config, &result)
if err != nil {
    log.Fatal(err)
}

// GET XML
var settings Config
err = client.GetXML("/config", &settings)
if err != nil {
    log.Fatal(err)
}
```

## Client Configuration

### Basic Configuration

```go
client := httpc.NewClient(
    httpc.WithBaseURL("https://api.example.com"),
    httpc.WithTimeout(30 * time.Second),
    httpc.WithHeader("User-Agent", "MyApp/1.0"),
    httpc.WithHeader("Accept", httpc.ContentTypeJSON),
)
```

### With Retry Logic

```go
retryConfig := httpc.DefaultRetryConfig()
retryConfig.MaxRetries = 3
retryConfig.Backoff = time.Second

client := httpc.NewClient(
    httpc.WithBaseURL("https://api.example.com"),
    httpc.WithRetry(retryConfig),
)
```

### With Interceptors

```go
client := httpc.NewClient(
    httpc.WithBaseURL("https://api.example.com"),
    httpc.WithDebug(),
    httpc.WithAuthorization("your-token"),
)
```

## Request Building

### Using Request Builder

```go
resp, err := client.NewRequest().
    Method("POST").
    URL("/users").
    Header("Content-Type", httpc.ContentTypeJSON).
    Query("filter", "active").
    JSON(map[string]string{"name": "John"}).
    Do()
```

### Using Convenience Methods

```go
// GET
resp, err := client.Get("/users",
    httpc.Header("Accept", httpc.ContentTypeJSON),
    httpc.WithQuery("page", "1"),
)

// POST
resp, err := client.Post("/users", userData)

// PUT
resp, err := client.Put("/users/123", userData)

// DELETE
resp, err := client.Delete("/users/123")
```

### With Context

All HTTP methods support context for cancellation, timeouts, and deadline propagation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// All methods have WithContext variants
resp, err := client.GetWithContext(ctx, "/users")
resp, err := client.PostWithContext(ctx, "/users", userData)
resp, err := client.PutWithContext(ctx, "/users/123", userData)
resp, err := client.DeleteWithContext(ctx, "/users/123")
resp, err := client.PatchWithContext(ctx, "/users/123", updates)

// JSON methods also support context
var users []User
err := client.GetJSONWithContext(ctx, "/users", &users)

user := User{Name: "Jane"}
var created User
err = client.PostJSONWithContext(ctx, "/users", user, &created)

// XML methods also support context
var config Config
err = client.GetXMLWithContext(ctx, "/config", &config)

settings := Config{Environment: "staging", Port: 9090}
var settingsResult Config
err = client.PostXMLWithContext(ctx, "/config", settings, &settingsResult)

// Or use WithContext as a RequestOption
resp, err := client.Get("/users",
    httpc.WithContext(ctx),
    httpc.WithQuery("page", "1"),
)
```

## Interceptors

Interceptors allow you to modify requests before they are sent. Several built-in interceptors are provided:

### Authentication

```go
// Bearer token
client := httpc.NewClient(
    httpc.WithAuthorization("your-token"),
)

// Basic auth
client := httpc.NewClient(
    httpc.WithBaseAuth("username", "password"),
)

// API Key
client := httpc.NewClient(
    httpc.WithApiKey("X-API-Key", "your-api-key"),
)
```

### User Agent

```go
// Set a custom User-Agent header for all requests
client := httpc.NewClient(
    httpc.WithUserAgent("MyApp/1.0"),
)
```

### Logging

```go
// With default logger
client := httpc.NewClient(
    httpc.WithLogger(log.Default()),
)

// Or with debug mode
client := httpc.NewClient(
    httpc.WithDebug(),
)
```

### Request ID

```go
client := httpc.NewClient(
    httpc.WithRequestId("X-Request-ID"),
)
```

### Custom Headers

```go
client := httpc.NewClient(
    httpc.WithHeaders(map[string]string{
        "X-Custom-Header": "value",
        "X-App-Version":   "1.0",
    }),
)
```

### Domain Blocking

```go
client := httpc.NewClient(
    httpc.WithBlockedList([]string{
        "blocked-domain.com",
        "another-blocked.com",
    }),
)
```

### Custom Interceptor

Interceptors wrap the underlying `http.RoundTripper` to add custom behavior:

```go
customInterceptor := func(rt http.RoundTripper) http.RoundTripper {
    return &customTransport{transport: rt}
}

type customTransport struct {
    transport http.RoundTripper
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
    // Modify request
    req = req.Clone(req.Context())
    req.Header.Set("X-Custom", "value")
    return t.transport.RoundTrip(req)
}

client := httpc.NewClient(
    httpc.WithInterceptor(customInterceptor),
)
```

### Dynamic Interceptors

You can also add interceptors after client creation:

```go
client := httpc.NewClient()
// Add logging interceptor dynamically
client.AddInterceptor(func(rt http.RoundTripper) http.RoundTripper {
    return &loggingTransport{
        transport: rt,
        logger:    log.Default(),
    }
})
```

## Response Handling

### As Bytes

```go
resp, err := client.Get("/users")
if err != nil {
    log.Fatal(err)
}

body, err := resp.Bytes()
if err != nil {
    log.Fatal(err)
}
```

### As String

```go
resp, err := client.Get("/users")
if err != nil {
    log.Fatal(err)
}

body, err := resp.String()
if err != nil {
    log.Fatal(err)
}
```

### As JSON

```go
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

resp, err := client.Get("/users/123")
if err != nil {
    log.Fatal(err)
}

var user User
err = resp.JSON(&user)
if err != nil {
    log.Fatal(err)
}
```

### As XML

```go
type Config struct {
    Environment string `xml:"environment"`
    Port        int    `xml:"port"`
}

resp, err := client.Get("/config")
if err != nil {
    log.Fatal(err)
}

var config Config
err = resp.XML(&config)
if err != nil {
    log.Fatal(err)
}
```

### Access Response Metadata

```go
resp, err := client.Get("/users")
if err != nil {
    log.Fatal(err)
}

fmt.Println("Status:", resp.StatusCode)
fmt.Println("Headers:", resp.Header)
fmt.Println("Content-Type:", resp.Header.Get("Content-Type"))
```

## Retry Configuration

### Basic Retry

```go
client := httpc.NewClient(
    httpc.WithRetry(httpc.DefaultRetryConfig()),
)
```

The default retry logic:
- Retries on any error
- Retries on 5xx status codes
- Retries on 429 (rate limit) status code
- Uses exponential backoff (backoff * attempt)

### Custom Retry Logic

```go
retryConfig := &httpc.RetryConfig{
    MaxRetries: 5,
    Backoff:    2 * time.Second,
    RetryIf: func(resp *http.Response, err error) bool {
        // Custom logic
        if err != nil {
            return true
        }
        // Only retry on specific status codes
        return resp.StatusCode == 503 || resp.StatusCode == 504
    },
}

client := httpc.NewClient(
    httpc.WithRetry(retryConfig),
)
```

## Error Handling

### HTTP Errors

```go
resp, err := client.Get("/users")
if err != nil {
    // Network error, timeout, etc.
    log.Fatal(err)
}

if !resp.isSuccess() { // Checks if status is 2xx
    // HTTP error (4xx, 5xx)
    body, _ := resp.String()
    log.Printf("HTTP error %d: %s", resp.StatusCode, body)
}
```

### Custom Error Type

```go
resp, err := client.Get("/users")
if err != nil {
    var httpErr *httpc.Error
    if errors.As(err, &httpErr) {
        log.Printf("Status: %d, Message: %s", httpErr.StatusCode, httpErr.Message)
    }
}
```

### Timeout Detection

```go
resp, err := client.Get("/users")
if err != nil {
    if httpc.IsTimeout(err) {
        log.Println("Request timed out")
    }
}
```

## Content Types

Pre-defined content type constants:

```go
httpc.ContentTypeJSON        // application/json
httpc.ContentTypeXML         // application/xml
httpc.ContentTypeForm        // application/x-www-form-urlencoded
httpc.ContentTypeMultipart   // multipart/form-data
httpc.ContentTypePlainText   // text/plain
httpc.ContentTypeHTML        // text/html
httpc.ContentTypeCSV         // text/csv
httpc.ContentTypeJavaScript  // application/javascript
httpc.ContentTypeCSS         // text/css
httpc.ContentTypePDF         // application/pdf
httpc.ContentTypeZip         // application/zip
httpc.ContentTypeOctetStream // application/octet-stream
```

## Advanced Examples

### Complete REST API Client

```go
type APIClient struct {
    client *httpc.Client
}

func NewAPIClient(baseURL, token string) *APIClient {
    retryConfig := httpc.DefaultRetryConfig()

    client := httpc.NewClient(
        httpc.WithBaseURL(baseURL),
        httpc.WithTimeout(30*time.Second),
        httpc.WithRetry(*retryConfig),
        httpc.WithAuthorization(token),
        httpc.WithDebug(),
        httpc.WithHeader("Accept", httpc.ContentTypeJSON),
    )

    return &APIClient{client: client}
}

func (a *APIClient) GetUser(id int) (*User, error) {
    var user User
    err := a.client.GetJSON(fmt.Sprintf("/users/%d", id), &user)
    return &user, err
}

func (a *APIClient) CreateUser(user *User) error {
    return a.client.PostJSON("/users", user, user)
}
```

### With Query Parameters

```go
resp, err := client.NewRequest().
    Method("GET").
    URL("/users").
    Query("page", "1").
    Query("limit", "10").
    Query("sort", "name").
    Do()

// Or with multiple parameters at once
resp, err := client.NewRequest().
    Method("GET").
    URL("/users").
    QueryParams(map[string]string{
        "page":  "1",
        "limit": "10",
        "sort":  "name",
    }).
    Do()
```

### With Request Timeout

The `Timeout()` method sets a request-specific timeout that overrides the client's default timeout. This creates a timeout context that will cancel the request if it exceeds the specified duration.

```go
resp, err := client.NewRequest().
    Method("GET").
    URL("/users").
    Timeout(5 * time.Second).  // Request times out after 5 seconds
    Do()

if httpc.IsTimeout(err) {
    log.Println("Request timed out")
}
```

**Timeout applies to:**
- Connection establishment
- Request sending
- Response reading
- The entire request/response cycle

**Combining Timeout with Context:**

```go
// You can combine both timeout and custom context
ctx := context.WithValue(context.Background(), "request-id", "123")
resp, err := client.NewRequest().
    Method("GET").
    URL("/users").
    Context(ctx).              // Custom context with values
    Timeout(5 * time.Second).  // Timeout is applied to this context
    Do()
```

### With Custom Body

```go
resp, err := client.NewRequest().
    Method("POST").
    URL("/upload").
    Header("Content-Type", "text/plain").
    Body(strings.NewReader("custom body content")).
    Do()
```

## Transport Configuration

The default transport is configured with:
- HTTP/2 support with fallback
- Connection pooling (100 max idle, 10 per host)
- 30s dial timeout
- 30s keep-alive
- 90s idle connection timeout
- 10s TLS handshake timeout
- Proxy from environment
- Compression disabled (for manual control)

## Thread Safety

The `Client` is safe for concurrent use. All methods are thread-safe and can be called from multiple goroutines simultaneously.

## Best Practices

1. **Reuse Clients**: Create one client and reuse it for multiple requests to benefit from connection pooling
2. **Use Context**: Always use context for cancellation and timeouts in production code
3. **Handle Errors**: Check both network errors and HTTP status codes
4. **Set Timeouts**: Always set appropriate timeouts to prevent hanging requests
5. **Use Interceptors**: Centralize cross-cutting concerns like auth, logging, and metrics
6. **Base URL**: Use `WithBaseURL` for APIs to avoid repeating the domain
7. **Retry Logic**: Enable retries for transient failures in production environments

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.
