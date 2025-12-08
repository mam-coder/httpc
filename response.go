package httpc

import (
	"encoding/json"
	"io"
	"net/http"
)

// Response wraps http.Response and provides convenience methods for
// reading and parsing the response body. The body is cached after the
// first read, so multiple calls to Bytes(), String(), or JSON() will
// return the same data without re-reading.
type Response struct {
	*http.Response
	body []byte
}

// Bytes return the response body as a byte slice.
// The body is read and cached on the first call, subsequent calls
// return the cached data without re-reading from the network.
//
// Example:
//
//	resp, err := client.Get("/api/users")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	body, err := resp.Bytes()
//	if err != nil {
//	    log.Fatal(err)
//	}
func (r *Response) Bytes() ([]byte, error) {
	if r.body != nil {
		return r.body, nil
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// Handle gzip-encoded responses
	if isGzipEncoded(r.Header.Get("Content-Encoding")) {
		body, err = decodeGzipBody(body)
		if err != nil {
			return nil, err
		}
	}

	r.body = body

	return body, nil
}

// String returns the response body as a string.
// The body is read and cached on the first call, subsequent calls
// return the cached data without re-reading from the network.
//
// Example:
//
//	resp, err := client.Get("/api/status")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	body, err := resp.String()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(body)
func (r *Response) String() (string, error) {
	body, err := r.Bytes()
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// JSON unmarshals the response body into the provided value.
// The body is read and cached on the first call.
//
// Example:
//
//	type User struct {
//	    ID   int    `json:"id"`
//	    Name string `json:"name"`
//	}
//
//	resp, err := client.Get("/api/users/123")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	var user User
//	if err := resp.JSON(&user); err != nil {
//	    log.Fatal(err)
//	}
func (r *Response) JSON(v interface{}) error {
	body, err := r.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// isSuccess returns true if the response status code is in the 2xx range.
func (r *Response) isSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}
