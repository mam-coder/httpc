package httpc

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Option is a function that configures a Client.
// Options are passed to NewClient to customize client behavior such as
// setting timeouts, base URLs, headers, retry logic, and interceptors.
//
// Example:
//
//	client := httpc.NewClient(
//	    httpc.WithBaseURL("https://api.example.com"),
//	    httpc.WithTimeout(30*time.Second),
//	    httpc.WithHeader("User-Agent", "MyApp/1.0"),
//	)
type Option func(*Client)

// WithBaseURL sets the base URL for all requests made by the client.
// Relative URLs in requests will be resolved against this base URL.
// If a request uses an absolute URL (starting with http:// or https://),
// the base URL is ignored for that request.
//
// Example:
//
//	client := httpc.NewClient(httpc.WithBaseURL("https://api.example.com"))
//	resp, err := client.Get("/users")  // Goes to https://api.example.com/users
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithTimeout sets the request timeout for all requests made by the client.
// Individual requests can override this timeout using the RequestBuilder.Timeout method.
// The default timeout is 30 seconds.
//
// Example:
//
//	client := httpc.NewClient(httpc.WithTimeout(10*time.Second))
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if c.httpClient != nil {
			c.httpClient.Timeout = timeout
		}
	}
}

// WithHeaders sets multiple HTTP headers that will be added to every request.
// Headers set here can be overridden on a per-request basis.
//
// Example:
//
//	client := httpc.NewClient(httpc.WithHeaders(map[string]string{
//	    "X-App-Version": "1.0",
//	    "X-Environment": "production",
//	}))
func WithHeaders(headers map[string]string) Option {
	return func(c *Client) {
		c.mu.Lock()
		defer c.mu.Unlock()
		for key, value := range headers {
			c.headers[key] = value
		}
	}
}

// WithHeader sets a single HTTP header that will be added to every request.
// Multiple calls to WithHeader will add multiple headers.
//
// Example:
//
//	client := httpc.NewClient(
//	    httpc.WithHeader("User-Agent", "MyApp/1.0"),
//	    httpc.WithHeader("Accept", "application/json"),
//	)
func WithHeader(key, value string) Option {
	return func(c *Client) {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.headers[key] = value
	}
}

// WithBaseAuth configures HTTP Basic Authentication for all requests.
// The username and password are automatically encoded and added to the
// Authorization header.
//
// Example:
//
//	client := httpc.NewClient(httpc.WithBaseAuth("user", "password"))
func WithBaseAuth(username, password string) Option {
	return WithInterceptor(func(rt http.RoundTripper) http.RoundTripper {
		return &baseAuthTransport{
			transport: rt,
			username:  username,
			password:  password,
		}
	})
}

// WithAuthorization sets a Bearer token for authentication.
// The token is added to the Authorization header as "Bearer <token>".
//
// Example:
//
//	client := httpc.NewClient(httpc.WithAuthorization("your-api-token"))
func WithAuthorization(token string) Option {
	return WithInterceptor(func(rt http.RoundTripper) http.RoundTripper {
		return &authTransport{
			transport: rt,
			token:     token,
		}
	})
}

// WithUserAgent sets the User-Agent header for all requests.
//
// Example:
//
//	client := httpc.NewClient(httpc.WithUserAgent("MyApp/1.0"))
func WithUserAgent(userAgent string) Option {
	return WithHeader("User-Agent", userAgent)
}

// WithContentType sets the Content-Type header for all requests.
//
// Example:
//
//	client := httpc.NewClient(httpc.WithContentType(httpc.ContentTypeJSON))
func WithContentType(contentType string) Option {
	return WithHeader("Content-Type", contentType)
}

// WithAccept sets the Accept header for all requests.
//
// Example:
//
//	client := httpc.NewClient(httpc.WithAccept(httpc.ContentTypeJSON))
func WithAccept(accept string) Option {
	return WithHeader("Accept", accept)
}

// WithApiKey sets an API key header for authentication.
// If headerName is empty, defaults to "X-Api-Key".
//
// Example:
//
//	client := httpc.NewClient(httpc.WithApiKey("X-API-Key", "secret-key-123"))
//	// Or with default header name:
//	client := httpc.NewClient(httpc.WithApiKey("", "secret-key-123"))
func WithApiKey(headerName, apiKey string) Option {
	if headerName == "" {
		headerName = "X-Api-Key"
	}
	return WithHeader(headerName, apiKey)
}

// WithRequestId generates and sets a unique request ID header for tracing.
// If headerName is empty, defaults to "X-Request-Id".
// A new request ID is generated for each client instance.
//
// Example:
//
//	client := httpc.NewClient(httpc.WithRequestId("X-Request-ID"))
func WithRequestId(headerName string) Option {
	if headerName == "" {
		headerName = "X-Request-Id"
	}
	requestID := fmt.Sprintf("req-%d", time.Now().UnixNano())
	return WithHeader(headerName, requestID)
}

// WithBlockedList configures a list of domains that should be blocked.
// Requests to any domain in the blockedList will fail with an error.
// This is useful for preventing requests to known malicious or unwanted domains.
//
// Example:
//
//	client := httpc.NewClient(httpc.WithBlockedList([]string{
//	    "malicious-site.com",
//	    "blocked-domain.com",
//	}))
func WithBlockedList(blockedList []string) Option {
	return WithInterceptor(func(rt http.RoundTripper) http.RoundTripper {
		return &blockListTransport{
			transport:   rt,
			blockedList: blockedList,
		}
	})
}

// WithRetry configures automatic retry logic with exponential backoff.
// Use RetryConfig to specify max retries, backoff duration, and retry conditions.
//
// Example:
//
//	config := httpc.RetryConfig{
//	    MaxRetries: 3,
//	    Backoff:    time.Second,
//	    RetryIf:    httpc.DefaultRetryCondition,  // or custom function
//	}
//	client := httpc.NewClient(httpc.WithRetry(config))
func WithRetry(config RetryConfig) Option {
	return WithInterceptor(func(rt http.RoundTripper) http.RoundTripper {
		return &retryTransport{
			transport: rt,
			config:    config,
		}
	})
}

// WithLogger configures request/response logging using the provided logger.
// All HTTP requests and responses will be logged with method, URL, status code, and timing.
//
// Example:
//
//	logger := log.New(os.Stdout, "[HTTP] ", log.LstdFlags)
//	client := httpc.NewClient(httpc.WithLogger(logger))
func WithLogger(logger *log.Logger) Option {
	return WithInterceptor(func(rt http.RoundTripper) http.RoundTripper {
		return &loggingTransport{
			transport: rt,
			logger:    logger,
		}
	})
}

// WithDebug enables debug logging for the client.
// When enabled, the client will log detailed information about requests and responses.
//
// Example:
//
//	client := httpc.NewClient(httpc.WithDebug())
func WithDebug() Option {
	return WithInterceptor(func(rt http.RoundTripper) http.RoundTripper {
		return NewDebugTransport(rt, true)
	})
}

// WithInterceptor adds a request interceptor to the client.
// Interceptors are executed in the order they are added and can modify
// requests before they are sent or return errors to prevent execution.
// Multiple interceptors can be added by calling this option multiple times.
//
// Example:
//
//	client := httpc.NewClient(
//	    httpc.WithInterceptor(httpc.LoggingInterceptor()),
//	    httpc.WithInterceptor(httpc.AuthInterceptor("token")),
//	)
func WithInterceptor(interceptor Interceptor) Option {
	return func(c *Client) {
		c.transport = interceptor(c.transport)
	}
}
