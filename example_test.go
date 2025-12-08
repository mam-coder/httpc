package httpc_test

import (
	"fmt"
	"log"
	"time"

	"github.com/mam-coder/httpc"
)

// Example demonstrates basic usage of the HTTP client.
func Example() {
	client := httpc.NewClient(
		httpc.WithBaseURL("https://api.example.com"),
		httpc.WithTimeout(10*time.Second),
	)

	resp, err := client.Get("/users")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Status:", resp.StatusCode)
}

// ExampleNewClient demonstrates creating a new HTTP client with options.
func ExampleNewClient() {
	client := httpc.NewClient(
		httpc.WithBaseURL("https://api.example.com"),
		httpc.WithTimeout(30*time.Second),
		httpc.WithHeader("User-Agent", "MyApp/1.0"),
	)

	resp, err := client.Get("/users")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Response received:", resp.StatusCode >= 200 && resp.StatusCode < 300)
}

// ExampleClient_Get demonstrates making a GET request.
func ExampleClient_Get() {
	client := httpc.NewClient(
		httpc.WithBaseURL("https://api.example.com"),
	)

	resp, err := client.Get("/users",
		httpc.WithQuery("page", "1"),
		httpc.WithQuery("limit", "10"),
	)
	if err != nil {
		log.Fatal(err)
	}

	body, _ := resp.String()
	fmt.Println("Response:", len(body) > 0)
}

// ExampleClient_Post demonstrates making a POST request with JSON.
func ExampleClient_Post() {
	client := httpc.NewClient(
		httpc.WithBaseURL("https://api.example.com"),
	)

	user := map[string]string{
		"name":  "John Doe",
		"email": "john@example.com",
	}

	resp, err := client.Post("/users", user)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Created:", resp.StatusCode == 201)
}

// ExampleClient_GetJSON demonstrates getting and parsing JSON response.
func ExampleClient_GetJSON() {
	client := httpc.NewClient(
		httpc.WithBaseURL("https://api.example.com"),
	)

	var users []map[string]interface{}
	err := client.GetJSON("/users", &users)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Got users:", len(users) >= 0)
}

// ExampleClient_NewRequest demonstrates using the request builder.
func ExampleClient_NewRequest() {
	client := httpc.NewClient()

	resp, err := client.NewRequest().
		Method("POST").
		URL("https://api.example.com/users").
		Header("Authorization", "Bearer token").
		Query("include", "profile").
		JSON(map[string]string{"name": "John"}).
		Timeout(5 * time.Second).
		Do()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Request completed:", resp != nil)
}

// ExampleWithRetry demonstrates configuring retry logic.
func ExampleWithRetry() {
	config := httpc.DefaultRetryConfig()
	config.MaxRetries = 3
	config.Backoff = time.Second

	client := httpc.NewClient(
		httpc.WithBaseURL("https://api.example.com"),
		httpc.WithRetry(*config),
	)

	resp, err := client.Get("/users")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Success:", resp.StatusCode == 200)
}

// ExampleWithAuthorization demonstrates Bearer token authentication.
func ExampleWithAuthorization() {
	client := httpc.NewClient(
		httpc.WithBaseURL("https://api.example.com"),
		httpc.WithAuthorization("your-api-token"),
	)

	resp, err := client.Get("/protected/resource")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Authenticated:", resp.StatusCode != 401)
}

// ExampleWithBaseAuth demonstrates HTTP Basic authentication.
func ExampleWithBaseAuth() {
	client := httpc.NewClient(
		httpc.WithBaseURL("https://api.example.com"),
		httpc.WithBaseAuth("username", "password"),
	)

	resp, err := client.Get("/protected/resource")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Authenticated:", resp.StatusCode != 401)
}

// ExampleIsTimeout demonstrates checking for timeout errors.
func ExampleIsTimeout() {
	client := httpc.NewClient(
		httpc.WithTimeout(1 * time.Millisecond),
	)

	_, err := client.Get("https://example.com/slow-endpoint")
	if httpc.IsTimeout(err) {
		fmt.Println("Request timed out")
	}
}