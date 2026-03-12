package httpclient

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	apiError "github.com/NamespaceManager/pkg/api_error"
)

// Client is a wrapper around http.Client with default timeout
var Client = &http.Client{
	Timeout: 30 * time.Second,
}

// MTLSClient is an HTTP client configured with client certificates for mutual TLS
var MTLSClient *http.Client

// InitMTLSClient initializes the mTLS client with client certificates
// This should be called during application startup if mTLS is enabled
func InitMTLSClient() error {
	// Load client certificate and key
	certPath := os.Getenv("TLS_CERT_PATH")
	keyPath := os.Getenv("TLS_KEY_PATH")
	caCertPath := os.Getenv("MTLS_CA_CERT_PATH")

	if certPath == "" || keyPath == "" {
		log.Println("mTLS client disabled: TLS_CERT_PATH or TLS_KEY_PATH not set")
		MTLSClient = Client // Fallback to regular client
		return nil
	}

	// Load client cert
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return err
	}

	// Load CA cert pool (for verifying server certificates)
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// If CA cert is provided, use it to verify server certificates
	if caCertPath != "" {
		caCert, err := os.ReadFile(caCertPath)
		if err != nil {
			log.Printf("Warning: Failed to read CA certificate: %v", err)
		} else {
			caCertPool := x509.NewCertPool()
			if caCertPool.AppendCertsFromPEM(caCert) {
				tlsConfig.RootCAs = caCertPool
			} else {
				log.Println("Warning: Failed to parse CA certificate")
			}
		}
	}

	MTLSClient = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	log.Println("mTLS client initialized successfully")
	return nil
}

// getClient returns the appropriate HTTP client
// Uses MTLSClient if available and initialized, otherwise falls back to regular Client
func getClient() *http.Client {
	if MTLSClient != nil {
		return MTLSClient
	}
	return Client
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

	// Execute request using the appropriate client (mTLS if available)
	client := getClient()
	resp, err := client.Do(req)
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

	// Execute request using the appropriate client (mTLS if available)
	client := getClient()
	resp, err := client.Do(req)
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
