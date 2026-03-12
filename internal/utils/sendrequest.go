package utils

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var client = &http.Client{Timeout: 10 * time.Second}
var mtlsClient *http.Client

// InitMTLSClient initializes the mTLS client for utils package
func InitMTLSClient() error {
	certPath := os.Getenv("TLS_CERT_PATH")
	keyPath := os.Getenv("TLS_KEY_PATH")
	caCertPath := os.Getenv("MTLS_CA_CERT_PATH")

	if certPath == "" || keyPath == "" {
		log.Println("utils mTLS client disabled: TLS_CERT_PATH or TLS_KEY_PATH not set")
		mtlsClient = client
		return nil
	}

	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

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

	mtlsClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	log.Println("utils mTLS client initialized successfully")
	return nil
}

// getClient returns the appropriate HTTP client
func getClient() *http.Client {
	if mtlsClient != nil {
		return mtlsClient
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
