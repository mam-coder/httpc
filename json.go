package httpc

import (
	"bytes"
	"context"
	"encoding/json"
)

// JSON marshals the provided value to JSON and sets it as the request body.
// It automatically sets the Content-Type header to application/json.
//
// Example:
//
//	client.NewRequest().
//	    Method("POST").
//	    URL("/api/users").
//	    JSON(map[string]string{"name": "John"}).
//	    Do()
func (rb *RequestBuilder) JSON(v interface{}) *RequestBuilder {
	data, err := json.Marshal(v)
	if err != nil {
		rb.err = err
	}

	rb.body = bytes.NewReader(data)
	rb.Header("Content-Type", ContentTypeJSON)

	return rb
}

// GetJSON is a convenience method that sends a GET request and automatically
// unmarshals the JSON response into the result parameter.
//
// Example:
//
//	var users []User
//	err := client.GetJSON("/api/users", &users,
//	    httpc.WithQuery("status", "active"),
//	)
func (c *Client) GetJSON(url string, result interface{}, opts ...RequestOption) error {
	resp, err := c.Get(url, opts...)
	if err != nil {
		return err
	}
	return resp.JSON(result)
}

// PostJSON is a convenience method that sends a POST request with a JSON body
// and automatically unmarshals the JSON response into the result parameter.
// Pass nil for result if you don't need to parse the response.
//
// Example:
//
//	user := User{Name: "John", Email: "john@example.com"}
//	var created User
//	err := client.PostJSON("/api/users", user, &created)
func (c *Client) PostJSON(url string, body interface{}, result interface{}, opts ...RequestOption) error {
	resp, err := c.Post(url, body, opts...)
	if err != nil {
		return err
	}
	if result != nil {
		return resp.JSON(result)
	}

	return nil
}

// GetJSONWithContext is a convenience method that sends a GET request with a context
// and automatically unmarshals the JSON response into the result parameter.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	var users []User
//	err := client.GetJSONWithContext(ctx, "/api/users", &users,
//	    httpc.WithQuery("status", "active"),
//	)
func (c *Client) GetJSONWithContext(ctx context.Context, url string, result interface{}, opts ...RequestOption) error {
	resp, err := c.GetWithContext(ctx, url, opts...)
	if err != nil {
		return err
	}
	return resp.JSON(result)
}

// PostJSONWithContext is a convenience method that sends a POST request with a context,
// JSON body, and automatically unmarshals the JSON response into the result parameter.
// Pass nil for result if you don't need to parse the response.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	user := User{Name: "John", Email: "john@example.com"}
//	var created User
//	err := client.PostJSONWithContext(ctx, "/api/users", user, &created)
func (c *Client) PostJSONWithContext(ctx context.Context, url string, body interface{}, result interface{}, opts ...RequestOption) error {
	resp, err := c.PostWithContext(ctx, url, body, opts...)
	if err != nil {
		return err
	}
	if result != nil {
		return resp.JSON(result)
	}

	return nil
}
