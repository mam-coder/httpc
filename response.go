package httpc

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"reflect"
)

// Response wraps http.Response and provides convenience methods for
// reading and parsing the response body. The body is cached after the
// first read, so multiple calls to Bytes(), String(), or JSON() will
// return the same data without re-reading.
type Response struct {
	*http.Response
	body         []byte
	csvSeparator rune
}

// Bytes returns the response body as a byte slice.
// The body is read and cached on the first call, subsequent calls
// return the cached data without re-reading from the network.
//
// Example:
//
//	resp, err := client.Get("/api/users")
//	if err != nil {
//		log.Fatal(err)
//	}
//	body, err := resp.Bytes()
//	if err != nil {
//		log.Fatal(err)
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
//		log.Fatal(err)
//	}
//	body, err := resp.String()
//	if err != nil {
//		log.Fatal(err)
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
//		ID   int    `json:"id"`
//		Name string `json:"name"`
//	}
//
//	resp, err := client.Get("/api/users/123")
//	if err != nil {
//		log.Fatal(err)
//	}
//	var user User
//	if err := resp.JSON(&user); err != nil {
//		log.Fatal(err)
//	}
func (r *Response) JSON(v interface{}) error {
	body, err := r.Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(body, v)
}

// XML unmarshals the response body into the provided value.
// The body is read and cached on the first call.
//
// Example:
//
//	type User struct {
//		ID   int    `xml:"id"`
//		Name string `xml:"name"`
//	}
func (r *Response) XML(v interface{}) error {
	body, err := r.Bytes()
	if err != nil {
		return err
	}
	return xml.Unmarshal(body, v)
}

// SetCSVSeparator sets the separator (delimiter) for CSV parsing.
// The default separator is comma (,). Use this method to parse
// TSV files (tab-separated) or other delimited formats.
//
// Example:
//
//	resp, err := client.Get("/api/users.tsv")
//	if err != nil {
//		log.Fatal(err)
//	}
//	resp.SetCSVSeparator('\t') // Set tab as separator
//	var users []User
//	if err := resp.CSV(&users); err != nil {
//		log.Fatal(err)
//	}
func (r *Response) SetCSVSeparator(sep rune) *Response {
	r.csvSeparator = sep
	return r
}

// CSV unmarshals the response body into the provided slice of structs.
// The first row is expected to contain header names that match struct field names or csv tags.
// The body is read and cached on the first call.
// By default, comma (,) is used as the separator. Use SetCSVSeparator to change it.
//
// Example:
//
//	type User struct {
//		ID   string `csv:"id"`
//		Name string `csv:"name"`
//	}
//
//	resp, err := client.Get("/api/users.csv")
//	if err != nil {
//		log.Fatal(err)
//	}
//	var users []User
//	if err := resp.CSV(&users); err != nil {
//		log.Fatal(err)
//	}
func (r *Response) CSV(v interface{}) error {
	body, err := r.Bytes()
	if err != nil {
		return err
	}

	reader := csv.NewReader(bytes.NewReader(body))

	// Set the separator if it was configured
	if r.csvSeparator != 0 {
		reader.Comma = r.csvSeparator
	}

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil
	}

	// Get the type of the slice
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("v must be a non-nil pointer to a slice")
	}

	sliceVal := rv.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return errors.New("v must be a pointer to a slice")
	}

	// Get the element type
	elemType := sliceVal.Type().Elem()

	// Parse header row
	headers := records[0]

	// Build a map from header name to column index
	headerMap := make(map[string]int)
	for i, header := range headers {
		headerMap[header] = i
	}

	// Process data rows
	for _, record := range records[1:] {
		// Create a new element
		elem := reflect.New(elemType).Elem()

		// Iterate over struct fields
		for i := 0; i < elemType.NumField(); i++ {
			field := elemType.Field(i)
			fieldValue := elem.Field(i)

			if !fieldValue.CanSet() {
				continue
			}

			// Get the CSV tag or use the field name
			csvTag := field.Tag.Get("csv")
			if csvTag == "" {
				csvTag = field.Name
			}

			// Find the column index
			colIndex, ok := headerMap[csvTag]
			if !ok || colIndex >= len(record) {
				continue
			}

			// Set the field value
			value := record[colIndex]
			if fieldValue.Kind() == reflect.String {
				fieldValue.SetString(value)
			}
		}

		// Append to the slice
		sliceVal.Set(reflect.Append(sliceVal, elem))
	}

	return nil
}

// isSuccess returns true if the response status code is in the 2xx range.
func (r *Response) isSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}
