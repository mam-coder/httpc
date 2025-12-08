package httpc

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"
)

// decodeGzipBody decodes gzip-encoded body data.
// It returns the decoded body or the original body if decoding fails.
func decodeGzipBody(body []byte) ([]byte, error) {
	if len(body) == 0 {
		return body, nil
	}

	reader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return body, err
	}
	defer reader.Close()

	decodedBody, err := io.ReadAll(reader)
	if err != nil {
		return body, err
	}

	return decodedBody, nil
}

// isGzipEncoded checks if the content encoding is gzip.
func isGzipEncoded(contentEncoding string) bool {
	return strings.ToLower(contentEncoding) == "gzip"
}
