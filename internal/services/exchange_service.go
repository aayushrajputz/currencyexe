package services

import (
	"fmt"
	"time"

	"exchange-rate-service/config"
)

// main service for currency ops
type CurrencyExchangeService struct {
	cache     ExchangeRateCache
	apiClient ExchangeRateAPIClient
}

// ExchangeRateCache defines what we need from our caching layer
type ExchangeRateCache interface {
	GetRate(fromCurrency, toCurrency string) (float64, bool)
	SetRate(fromCurrency, toCurrency string, rate float64)
}

// ExchangeRateAPIClient defines what we need from our API client
type ExchangeRateAPIClient interface {
	GetRate(fromCurrency, toCurrency, dateStr string) (float64, error)
}

// create new service
func NewCurrencyExchangeService(cache ExchangeRateCache, apiClient ExchangeRateAPIClient) *CurrencyExchangeService {
	return &CurrencyExchangeService{
		cache:     cache,
		apiClient: apiClient,
	}
}

// convert currency amount
func (s *CurrencyExchangeService) ConvertCurrencyAmount(from, to string, amt float64, dt string) (float64, error) {
	// validate inputs
	if err := s.validateCurrencyPair(from, to); err != nil {
		return 0, err
	}

	if amt < 0 {
		return 0, fmt.Errorf("amount cannot be negative: %f", amt)
	}

	// same currency = no conversion needed
	if from == to {
		return amt, nil
	}

	// get rate for this pair
	rate, err := s.getExchangeRateForPair(from, to, dt)
	if err != nil {
		return 0, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	result := amt * rate
	return result, nil
}

// GetHistoricalRate retrieves historical exchange rate for a specific date
func (service *CurrencyExchangeService) GetHistoricalExchangeRate(fromCurrency, toCurrency, dateStr string) (float64, error) {
	// Validate the currency pair first
	if err := service.validateCurrencyPair(fromCurrency, toCurrency); err != nil {
		return 0, err
	}

	// Same currency is always 1:1, even historically
	if fromCurrency == toCurrency {
		return 1.0, nil
	}

	// Parse and validate the date
	parsedDate, err := service.validateAndParseDate(dateStr)
	if err != nil {
		return 0, err
	}

	// Check if the date is within our allowed historical range
	if err := service.validateHistoricalRange(parsedDate); err != nil {
		return 0, err
	}

	// get historical rate - no caching for historical data
	historicalRate, err := service.apiClient.GetRate(fromCurrency, toCurrency, dateStr)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch historical rate: %w", err)
	}

	return historicalRate, nil
}

// getExchangeRateForPair retrieves exchange rate, using cache for latest rates

func (service *CurrencyExchangeService) getExchangeRateForPair(fromCurrency, toCurrency, dateStr string) (float64, error) {
	// For historical dates, we always fetch fresh from the API (no caching)
	if dateStr != "" {
		parsedDate, err := service.validateAndParseDate(dateStr)
		if err != nil {
			return 0, err
		}

		if err := service.validateHistoricalRange(parsedDate); err != nil {
			return 0, err
		}

		return service.apiClient.GetRate(fromCurrency, toCurrency, dateStr)
	}

	// check cache first
	if rate, found := service.cache.GetRate(fromCurrency, toCurrency); found {
		return rate, nil
	}

	// cache miss - fetch from api
	rate, err := service.apiClient.GetRate(fromCurrency, toCurrency, "")
	if err != nil {
		return 0, err
	}

	// cache the result
	service.cache.SetRate(fromCurrency, toCurrency, rate)

	return rate, nil
}

// validateCurrencies checks if both currencies are supported
func (service *CurrencyExchangeService) validateCurrencyPair(fromCurrency, toCurrency string) error {
	if !config.IsSupportedCurrency(fromCurrency) {
		return fmt.Errorf("unsupported source currency: %s", fromCurrency)
	}

	if !config.IsSupportedCurrency(toCurrency) {
		return fmt.Errorf("unsupported target currency: %s", toCurrency)
	}

	return nil
}

// validateAndParseDate validates date format and parses it
func (service *CurrencyExchangeService) validateAndParseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date cannot be empty")
	}

	// Parse the date string using the standard ISO format (YYYY-MM-DD)
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %s", dateStr)
	}

	// Don't allow future dates - that doesn't make business sense
	if parsedDate.After(time.Now()) {
		return time.Time{}, fmt.Errorf("date cannot be in the future: %s", dateStr)
	}

	return parsedDate, nil
}

// validateHistoricalRange checks if the date is within allowed historical range
func (service *CurrencyExchangeService) validateHistoricalRange(requestedDate time.Time) error {
	// Calculate the oldest date we allow based on our business rules
	oldestAllowedDate := time.Now().AddDate(0, 0, -config.MaxHistoricalDays)

	if requestedDate.Before(oldestAllowedDate) {
		return fmt.Errorf("date is too far in the past, maximum %d days allowed", config.MaxHistoricalDays)
	}

	return nil
}
