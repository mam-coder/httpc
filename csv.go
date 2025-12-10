package httpc

import "context"

// GetCSV is a convenience method that sends a GET request and automatically
// unmarshals the CSV response into the result parameter.
// The result must be a pointer to a slice of structs.
//
// Example:
//
//	type User struct {
//		ID   string `csv:"id"`
//		Name string `csv:"name"`
//	}
//
//	var users []User
//	err := client.GetCSV("/api/users.csv", &users,
//	    httpc.WithQuery("status", "active"),
//	)
func (c *Client) GetCSV(url string, result interface{}, opts ...RequestOption) error {
	resp, err := c.Get(url, opts...)
	if err != nil {
		return err
	}
	return resp.CSV(result)
}

// GetCSVWithContext is a convenience method that sends a GET request with a context
// and automatically unmarshals the CSV response into the result parameter.
// The result must be a pointer to a slice of structs.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	type User struct {
//		ID   string `csv:"id"`
//		Name string `csv:"name"`
//	}
//
//	var users []User
//	err := client.GetCSVWithContext(ctx, "/api/users.csv", &users,
//	    httpc.WithQuery("status", "active"),
//	)
func (c *Client) GetCSVWithContext(ctx context.Context, url string, result interface{}, opts ...RequestOption) error {
	resp, err := c.GetWithContext(ctx, url, opts...)
	if err != nil {
		return err
	}
	return resp.CSV(result)
}

// GetCSVWithSeparator is a convenience method that sends a GET request and automatically
// unmarshals the CSV response using a custom separator into the result parameter.
// This is useful for TSV files or other delimited formats.
// The result must be a pointer to a slice of structs.
//
// Example:
//
//	type User struct {
//		ID   string `csv:"id"`
//		Name string `csv:"name"`
//	}
//
//	var users []User
//	err := client.GetCSVWithSeparator("/api/users.tsv", '\t', &users,
//	    httpc.WithQuery("status", "active"),
//	)
func (c *Client) GetCSVWithSeparator(url string, separator rune, result interface{}, opts ...RequestOption) error {
	resp, err := c.Get(url, opts...)
	if err != nil {
		return err
	}
	return resp.SetCSVSeparator(separator).CSV(result)
}

// GetCSVWithSeparatorAndContext is a convenience method that sends a GET request with a context
// and automatically unmarshals the CSV response using a custom separator into the result parameter.
// This is useful for TSV files or other delimited formats.
// The result must be a pointer to a slice of structs.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	type User struct {
//		ID   string `csv:"id"`
//		Name string `csv:"name"`
//	}
//
//	var users []User
//	err := client.GetCSVWithSeparatorAndContext(ctx, "/api/users.tsv", '\t', &users,
//	    httpc.WithQuery("status", "active"),
//	)
func (c *Client) GetCSVWithSeparatorAndContext(ctx context.Context, url string, separator rune, result interface{}, opts ...RequestOption) error {
	resp, err := c.GetWithContext(ctx, url, opts...)
	if err != nil {
		return err
	}
	return resp.SetCSVSeparator(separator).CSV(result)
}
