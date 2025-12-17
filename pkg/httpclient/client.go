package httpclient

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	apiError "github.com/NamespaceManager/pkg/api_error"
)

// Client is a wrapper around http.Client with default timeout
var Client = &http.Client{
	Timeout: 10 * time.Second,
}

// SendRequest sends an HTTP request without authentication
// Returns response body on success (2xx status), otherwise returns appropriate ApiError
func SendRequest(url string, payload interface{}, method string) ([]byte, error) {
	// Prepare request body
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, apiError.NewBadRequestError("failed to marshal request payload")
		}
		body = bytes.NewBuffer(jsonData)
	}

	// Create HTTP request
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, apiError.NewInternalServerError("failed to create HTTP request")
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := Client.Do(req)
	if err != nil {
		return nil, apiError.NewServiceUnavailableError("failed to connect to external service")
	}
	defer resp.Body.Close()

	log.Println("request url", url)
	log.Println("request method", method)
	log.Println("response status", resp.StatusCode)
	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apiError.NewInternalServerError("failed to read response body")
	}

	// Handle response status
	return handleResponse(resp.StatusCode, responseBody)
}

// SendRequestWithAccessToken sends an HTTP request with optional Bearer token authentication
// Returns response body on success (2xx status), otherwise returns appropriate ApiError
func SendRequestWithAccessToken(url string, payload interface{}, method string, accessToken string) ([]byte, error) {
	// Prepare request body
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, apiError.NewBadRequestError("failed to marshal request payload")
		}
		body = bytes.NewBuffer(jsonData)
	}

	// Create HTTP request
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, apiError.NewInternalServerError("failed to create HTTP request")
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	// Execute request
	resp, err := Client.Do(req)
	if err != nil {
		return nil, apiError.NewServiceUnavailableError("failed to connect to external service")
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apiError.NewInternalServerError("failed to read response body")
	}

	// Handle response status
	return handleResponse(resp.StatusCode, responseBody)
}
