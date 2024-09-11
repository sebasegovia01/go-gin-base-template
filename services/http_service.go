package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type HTTPServiceInterface interface {
	SendRequest(method, url string, params map[string]string) ([]byte, error)
}

type HTTPService struct {
	client *http.Client
}

// NewHTTPService initializes the HTTP service
func NewHTTPService() *HTTPService {
	return &HTTPService{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendRequest sends an HTTP request with optional query parameters
func (h *HTTPService) SendRequest(method, url string, params map[string]string) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Adding query parameters if they exist
	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	// Log the request URL
	log.Printf("Sending %s request to: %s", method, req.URL.String())

	// Send the request
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read and return the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response: %s", resp.Status)
	}

	return body, nil
}
