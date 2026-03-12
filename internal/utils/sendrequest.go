package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/NamespaceManager/pkg/httpclient"
)

var client = &http.Client{Timeout: 10 * time.Second}

// getClient returns the appropriate HTTP client
// Uses httpclient.MTLSClient if available (mTLS-enabled), otherwise falls back to regular client
func getClient() *http.Client {
	if httpclient.MTLSClient != nil {
		return httpclient.MTLSClient
	}
	return client
}

func SendRequest(url string, payload interface{}, method string) (int, []byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// sessionCookie := os.Getenv("SESSION_COOKIE")

	// if sessionCookie != "" {
	// 	req.Header.Set("Cookie", sessionCookie)
	// }

	httpClient := getClient()
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("failed to read response: %w", err)
	}

	// bodyString := string(body)

	return resp.StatusCode, body, nil
}
