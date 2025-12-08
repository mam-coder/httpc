// Package httpc provides HTTP client functionality.
// This file defines the Interceptor type for wrapping http.RoundTripper implementations.
package httpc

import (
	"net/http"
)

// Interceptor is a function that wraps an http.RoundTripper to add custom behavior.
// Interceptors can be chained together to create complex request/response processing pipelines.
//
// Example:
//
//	customInterceptor := func(rt http.RoundTripper) http.RoundTripper {
//	    return &myCustomTransport{transport: rt}
//	}
//	client := httpc.NewClient(httpc.WithInterceptor(customInterceptor))
type Interceptor func(tripper http.RoundTripper) http.RoundTripper
