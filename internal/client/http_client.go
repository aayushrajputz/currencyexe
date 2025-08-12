package client

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HTTPClient wraps the standard HTTP client with additional functionality
type HTTPClient struct {
	client  *http.Client
	baseURL string
	headers map[string]string
}

// NewHTTPClient creates a new HTTP client with sensible defaults
func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		baseURL: baseURL,
		headers: make(map[string]string),
	}
}

// SetHeader sets a default header for all requests
func (c *HTTPClient) SetHeader(key, value string) {
	c.headers[key] = value
}

// Get performs a GET request
func (c *HTTPClient) Get(ctx context.Context, endpoint string) (*http.Response, error) {
	return c.doRequest(ctx, "GET", endpoint, nil)
}

// doRequest performs the actual HTTP request with common setup
func (c *HTTPClient) doRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add default headers
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}
	
	// Set common headers
	req.Header.Set("User-Agent", "exchange-rate-service/1.0.0")
	req.Header.Set("Accept", "application/json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	
	return resp, nil
}

// Close cleans up the HTTP client resources
func (c *HTTPClient) Close() {
	c.client.CloseIdleConnections()
}
