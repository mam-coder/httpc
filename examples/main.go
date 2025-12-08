package main

import (
	"fmt"
	"log"

	"github.com/mam-coder/httpc"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	client := httpc.NewClient(
		httpc.WithBaseURL(
			"http://api.example.com"),
		httpc.WithHeader("Accept", httpc.ContentTypeJSON),
		httpc.WithHeader("Accept-Encoding", "gzip, deflate, br"),
		httpc.WithHeader("Content-Type", "application/json; charsets=utf-8"),
		httpc.WithUserAgent("MyApp/v1.0.5"),
		//httpc.WithBlockedList([]string{"api.bookair.zeit.test"}),
		httpc.WithDebug(),
		httpc.WithLogger(log.Default()),
	)

	resp, err := client.Get("/v1/status/health")
	// Check what encoding was used

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Status)
	fmt.Println("Content-Type:", resp.Header.Get("Content-Type"))

	body, err := resp.Bytes()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(body))
}
