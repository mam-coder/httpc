# CHANGELOG

## Transport-Based Architecture Refactor

### Summary
The httpc package has been refactored to use a transport-based architecture with `http.RoundTripper` wrappers instead of the previous interceptor-based approach. This provides better composability, performance, and follows Go's standard library patterns.

### Breaking Changes

#### Interceptor API Changed
**Before:**
```go
type Interceptor interface {
    Intercept(*http.Request) error
}
```

**After:**
```go
type Interceptor func(http.RoundTripper) http.RoundTripper
```

Interceptors now wrap the underlying transport instead of modifying requests directly.

#### Retry Configuration Changed
**Before:**
```go
client := httpc.NewClient(
    httpc.WithRetries(3, time.Second),
)
```

**After:**
```go
retryConfig := httpc.DefaultRetryConfig()
retryConfig.MaxRetries = 3
retryConfig.Backoff = time.Second

client := httpc.NewClient(
    httpc.WithRetry(*retryConfig),
)
```

#### Built-in Interceptors Replaced with Options

Several built-in interceptor functions have been replaced with cleaner option functions:

| Old API | New API |
|---------|---------|
| `AuthInterceptor(token)` | `WithAuthorization(token)` |
| `BasicAuthInterceptor(user, pass)` | `WithBaseAuth(user, pass)` |
| `APIKeyInterceptor(header, key)` | `WithApiKey(header, key)` |
| `LoggingInterceptor()` | `WithLogger(logger)` or `WithDebug()` |
| `UserAgentInterceptor(ua)` | `WithUserAgent(ua)` |
| `CustomHeaderInterceptor(headers)` | `WithHeaders(headers)` |
| `BlockListInterceptor(domains)` | `WithBlockedList(domains)` |
| `RequestIDInterceptor()` | `WithRequestId(header)` |
| `RateLimitInterceptor(rps)` | **Removed** (TODO) |
| `ConditionalAuthInterceptor(...)` | **Removed** |
| `ValidationInterceptor()` | **Removed** |

### New Features

#### Transport Wrappers
New internal transport types for better modularity:
- `headerTransport` - Adds custom headers
- `authTransport` - Bearer token authentication
- `baseAuthTransport` - HTTP Basic authentication
- `blockListTransport` - Domain blocking
- `retryTransport` - Retry logic with exponential backoff
- `loggingTransport` - Request/response logging

#### Improved Options
- `WithDebug()` - Enable debug mode with detailed logging
- `WithLogger(*log.Logger)` - Custom logger configuration
- `WithRetry(RetryConfig)` - Flexible retry configuration
- `WithBlockedList([]string)` - Domain blocking

### Files Updated

#### Documentation Added
✅ **client.go** - Added package doc and documented Client type and NewClient function
✅ **debug.go** - Added package doc for DebugTransport implementation
✅ **interceptor.go** - Added package doc and documented Interceptor type
✅ **transport.go** - Added package doc for transport implementations
✅ **options.go** - Updated all option functions with new transport-based approach
✅ **retry.go** - Documented RetryConfig and retry behavior

#### Tests Updated
✅ **client_test.go** - Updated for transport-based architecture
✅ **options_test.go** - Updated for new options API
✅ **interceptor_test.go** - Completely rewritten for transport wrappers
✅ **retry_test.go** - Updated for WithRetry(RetryConfig) API
✅ **request_test.go** - Updated header access tests
✅ **debug_test.go** - New comprehensive test file (20+ tests)

#### Files Removed
❌ **example_test.go** - Removed (examples are comprehensive in README.md)

## Test Coverage

All test files now have descriptive package documentation:
- client_test.go - Client initialization and configuration tests
- options_test.go - Configuration option tests
- request_test.go - Request builder and URL resolution tests
- retry_test.go - Retry logic and backoff tests
- interceptor_test.go - Transport wrapper tests
- debug_test.go - Debug transport tests
- errors_test.go - Error handling tests
- response_test.go - Response handling tests
- json_test.go - JSON marshaling tests
- methods_test.go - HTTP method convenience function tests

## Test Results
```bash
$ go test ./...
ok  	github.com/mam-coder/httpc	5.672s
```

✅ **All tests passing**

## Documentation Quality

### Package-level docs
Every .go file now has proper package-level documentation explaining its purpose.

### Type documentation
All exported types are documented:
- Client
- DebugTransport
- Error
- Interceptor
- RequestOption
- Option
- RequestBuilder
- Response
- RetryConfig

### Function documentation
All exported functions have documentation with examples where appropriate.

### godoc Output
Run `go doc -all httpc` to see complete generated documentation.

## Files With Documentation

| File | Package Doc | Type Docs | Function Docs |
|------|-------------|-----------|---------------|
| client.go | ✅ | ✅ | ✅ |
| debug.go | ✅ | ✅ | ✅ |
| doc.go | ✅ | N/A | N/A |
| errors.go | ✅ | ✅ | ✅ |
| interceptor.go | ✅ | ✅ | N/A |
| json.go | ✅ | N/A | ✅ |
| methods.go | ✅ | ✅ | ✅ |
| options.go | ✅ | ✅ | ✅ |
| request.go | ✅ | ✅ | ✅ |
| response.go | ✅ | ✅ | ✅ |
| retry.go | ✅ | ✅ | ✅ |
| transport.go | ✅ | N/A | ✅ |
| const.go | ✅ | N/A | N/A |

## Next Steps

The codebase is now fully documented and tested. All files have:
1. ✅ Package-level documentation
2. ✅ Type documentation for exported types
3. ✅ Function documentation for exported functions
4. ✅ Comprehensive test coverage
5. ✅ All tests passing

The project is ready for:
- Publishing to pkg.go.dev
- Code review
- Production use
