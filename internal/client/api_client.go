package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"exchange-rate-service/config"
)

// RateClient wraps http calls to exchange api
type RateClient struct {
	client  *HTTPClient
	baseurl string
}

// NewRateClient init new client
func NewRateClient() *RateClient {
	timeout := config.DefaultAPITimeout
	httpclient := NewHTTPClient(config.ExternalAPIBaseURL, timeout)

	return &RateClient{
		client:  httpclient,
		baseurl: config.ExternalAPIBaseURL,
	}
}

// apiResp from exchangerate-api.com
type apiResp struct {
	Result             string  `json:"result"`
	Documentation      string  `json:"documentation"`
	TermsOfUse         string  `json:"terms_of_use"`
	TimeLastUpdateUnix int64   `json:"time_last_update_unix"`
	TimeLastUpdateUTC  string  `json:"time_last_update_utc"`
	TimeNextUpdateUnix int64   `json:"time_next_update_unix"`
	TimeNextUpdateUTC  string  `json:"time_next_update_utc"`
	BaseCode           string  `json:"base_code"`
	TargetCode         string  `json:"target_code"`
	ConversionRate     float64 `json:"conversion_rate"`
	ConversionResult   float64 `json:"conversion_result"`
}

// GetRate gets exchange rate with retry
func (c *RateClient) GetRate(from, to, date string) (float64, error) {
	maxRetries := 2
	retryDelay := 500

	var lastErr error

	for i := 1; i <= maxRetries; i++ {
		rate, err := c.doAPICall(from, to, date)
		if err == nil {
			return rate, nil
		}

		lastErr = err

		if i < maxRetries {
			time.Sleep(time.Duration(retryDelay) * time.Millisecond)
		}
	}

	return 0, fmt.Errorf("failed after %d tries: %w", maxRetries, lastErr)
}

// doAPICall single http req
func (c *RateClient) doAPICall(from, to, dt string) (float64, error) {
	timeout := 12 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	endpoint := c.buildEndpoint(from, to, dt)

	resp, err := c.client.Get(ctx, endpoint)
	if err != nil {
		return 0, fmt.Errorf("http req failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("api http %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read body failed: %w", err)
	}

	var response apiResp
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("json parse failed: %w", err)
	}

	if response.Result != "success" {
		return 0, fmt.Errorf("api error: %s", response.Result)
	}

	if response.ConversionRate <= 0 {
		return 0, fmt.Errorf("invalid rate: %f", response.ConversionRate)
	}

	return response.ConversionRate, nil
}

// buildEndpoint makes url path
func (c *RateClient) buildEndpoint(from, to, dt string) string {
	// ignore date for now - need paid plan
	return fmt.Sprintf("/%s/pair/%s/%s/1", config.ExchangeRateAPIKey, from, to)
}

// Close cleanup
func (c *RateClient) Close() {
	c.client.Close()
}
