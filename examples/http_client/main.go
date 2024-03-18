package main

import (
	"bitbucket.org/rctiplus/vegapunk/http_client"
	"bitbucket.org/rctiplus/vegapunk/log"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

func parseHTTPResponse(resp *http.Response) (string, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func basicHTTPRequest(ctx context.Context, req *http.Request) {
	client := http_client.NewHttpClient(nil)
	resp, err := client.Do(ctx, req)
	fmt.Println("\nBasic HTTP Request")
	if err != nil {
		log.Errorf("error making request: %v\n", err)
		return
	}

	body, _ := parseHTTPResponse(resp)
	fmt.Printf("Joke: %s\n", body)
}

func basicHTTPRequestWithTimeout(ctx context.Context, req *http.Request) {
	client := http_client.NewHttpClientWithTimeout(http_client.NewHttpClient(nil), 500, 10, 10)
	resp, err := client.Do(ctx, req)
	fmt.Println("\nBasic HTTP Request With Timeout")
	if err != nil {
		log.Errorf("error making request: %v\n", err)
		return
	}

	body, _ := parseHTTPResponse(resp)
	fmt.Printf("Joke: %s\n", body)
}

func basicHTTPRequestWithSingleFlight(ctx context.Context, req *http.Request) {
	client := http_client.NewHttpWithSingleFlight(http_client.NewHttpClient(nil))
	resp, err := client.Do(ctx, req)
	fmt.Println("\nBasic HTTP Request With SingleFlight")
	if err != nil {
		log.Errorf("error making request: %v\n", err)
		return
	}

	body, _ := parseHTTPResponse(resp)
	fmt.Printf("Joke: %s\n", body)
}

func main() {
	ctx := context.Background()

	url := "https://icanhazdadjoke.com/"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "text/plain")

	go basicHTTPRequest(ctx, req)
	go basicHTTPRequestWithTimeout(ctx, req)
	go basicHTTPRequestWithSingleFlight(ctx, req)

	time.Sleep(1 * time.Second)
}
