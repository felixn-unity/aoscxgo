package aoscxgo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// RequestError represents a custom error for HTTP requests
type RequestError struct {
	StatusCode string
	Err        error
}

// Error implements the error interface
func (r *RequestError) Error() string {
	return fmt.Sprintf("Status: %s, Error: %v", r.StatusCode, r.Err)
}

// setupRequest creates and configures a basic HTTP request with common headers
func setupRequest(client *Client, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "*/*")
	req.Header.Set("x-csrf-token", client.Csrf)
	req.Close = false
	req.AddCookie(client.Cookie)

	return req, nil
}

// executeRequest performs the HTTP request and handles common errors
func executeRequest(client *Client, req *http.Request) *http.Response {
	res, err := client.Transport.RoundTrip(req)
	if err != nil {
		log.Fatalf("An Error Occurred %v", err)
	}
	return res
}

// delete performs DELETE to the given URL and returns the response
func delete(client *Client, url string) *http.Response {
	req, err := setupRequest(client, "DELETE", url, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	return executeRequest(client, req)
}

// get performs GET to the given URL and returns the response and parsed JSON body
func get(client *Client, url string) (*http.Response, map[string]interface{}) {
	req, err := setupRequest(client, "GET", url, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "text/plain")

	res := executeRequest(client, req)

	body := make(map[string]interface{})

	// Check if response is a server error (5xx) - these are unexpected
	if res.StatusCode >= 500 {
		log.Printf("HTTP Server Error %d: %s", res.StatusCode, res.Status)
		return res, body
	}

	// For client errors (4xx), just return empty body - these are often expected (like 404 for existence checks)
	if res.StatusCode >= 400 {
		return res, body
	}

	// Read the response body first
	bodyBytes, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return res, body
	}

	// Check if the response looks like JSON
	bodyStr := string(bodyBytes)
	if len(bodyStr) == 0 {
		return res, body
	}

	// Check content type to see if it's JSON
	contentType := res.Header.Get("Content-Type")
	if contentType != "" && !contains(contentType, "application/json") && !contains(contentType, "text/json") {
		log.Printf("Warning: Response content-type is '%s', not JSON\nResponse body: %s", contentType, bodyStr)
		return res, body
	}

	// Try to decode JSON from the bytes
	if err := json.Unmarshal(bodyBytes, &body); err != nil {
		log.Printf("Failed to decode JSON response: %v\nResponse body: %s", err, bodyStr)
		// Return empty body map instead of failing
		return res, make(map[string]interface{})
	}

	return res, body
}

// contains checks if a string contains a substring (case-insensitive helper)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				indexOf(s, substr) >= 0)))
}

// indexOf finds the index of substr in s
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// getAcceptText performs GET to the given URL with text/plain Accept header
func getAcceptText(client *Client, url string) (*http.Response, error) {
	req, err := setupRequest(client, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/plain")

	res, err := client.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// post performs POST to the given URL with the provided body and returns the response
func post(client *Client, url string, jsonBody *bytes.Buffer) *http.Response {
	req, err := setupRequest(client, "POST", url, jsonBody)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return executeRequest(client, req)
}

// put performs PUT to the given URL with the provided body and returns the response
func put(client *Client, url string, jsonBody *bytes.Buffer) *http.Response {
	req, err := setupRequest(client, "PUT", url, jsonBody)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return executeRequest(client, req)
}

// patch performs PATCH to the given URL with the provided body and returns the response
func patch(client *Client, url string, jsonBody *bytes.Buffer) *http.Response {
	req, err := setupRequest(client, "PATCH", url, jsonBody)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return executeRequest(client, req)
}
