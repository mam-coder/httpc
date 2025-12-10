package httpc

import (
	"bytes"
	"context"
	"encoding/xml"
)

// XML marshals the provided value to XML and sets it as the request body.
// It automatically sets the Content-Type header to application/xml.
//
// Example:
//
//	client.NewRequest().
//	    Method("POST").
//	    URL("/api/config").
//	    XML(map[string]string{"key": "value"}).
//	    Do()
func (rb *RequestBuilder) XML(v interface{}) *RequestBuilder {
	data, err := xml.Marshal(v)
	if err != nil {
		rb.err = err
	}

	rb.body = bytes.NewReader(data)
	rb.Header("Content-Type", ContentTypeXML)

	return rb
}

// GetXML is a convenience method that sends a GET request and automatically
// unmarshals the XML response into the result parameter.
//
// Example:
//
//	var config Config
//	err := client.GetXML("/api/config", &config,
//	    httpc.WithQuery("env", "production"),
//	)
func (c *Client) GetXML(url string, result interface{}, opts ...RequestOption) error {
	resp, err := c.Get(url, opts...)
	if err != nil {
		return err
	}
	return resp.XML(result)
}

// PostXML is a convenience method that sends a POST request with an XML body
// and automatically unmarshals the XML response into the result parameter.
// Pass nil for result if you don't need to parse the response.
//
// Example:
//
//	config := Config{Environment: "production"}
//	var created Config
//	err := client.PostXML("/api/config", config, &created)
func (c *Client) PostXML(url string, body interface{}, result interface{}, opts ...RequestOption) error {
	rb := c.NewRequest().Method("POST").URL(url)

	if body != nil {
		rb.XML(body)
	}

	for _, opt := range opts {
		opt(rb)
	}

	resp, err := rb.Do()
	if err != nil {
		return err
	}
	if result != nil {
		return resp.XML(result)
	}

	return nil
}

// GetXMLWithContext is a convenience method that sends a GET request with a context
// and automatically unmarshals the XML response into the result parameter.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	var config Config
//	err := client.GetXMLWithContext(ctx, "/api/config", &config,
//	    httpc.WithQuery("env", "production"),
//	)
func (c *Client) GetXMLWithContext(ctx context.Context, url string, result interface{}, opts ...RequestOption) error {
	resp, err := c.GetWithContext(ctx, url, opts...)
	if err != nil {
		return err
	}
	return resp.XML(result)
}

// PostXMLWithContext is a convenience method that sends a POST request with a context,
// XML body, and automatically unmarshals the XML response into the result parameter.
// Pass nil for result if you don't need to parse the response.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	config := Config{Environment: "production"}
//	var created Config
//	err := client.PostXMLWithContext(ctx, "/api/config", config, &created)
func (c *Client) PostXMLWithContext(ctx context.Context, url string, body interface{}, result interface{}, opts ...RequestOption) error {
	rb := c.NewRequest().Method("POST").URL(url).Context(ctx)

	if body != nil {
		rb.XML(body)
	}

	for _, opt := range opts {
		opt(rb)
	}

	resp, err := rb.Do()
	if err != nil {
		return err
	}
	if result != nil {
		return resp.XML(result)
	}

	return nil
}
