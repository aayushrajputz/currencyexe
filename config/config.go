package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// app constants
const (
	DefaultServerPort     = "8080"
	MaxAllowedHistoryDays = 90
	CacheRefreshInterval  = time.Hour
	DefaultAPITimeout     = 15 * time.Second
)

// supported currencies
// todo: move to db?
var SupportedCurrencyList = []string{"USD", "INR", "EUR", "JPY", "GBP"}

// Global config variables - loaded once at startup
var (
	ExternalAPIBaseURL string
	ExchangeRateAPIKey string
	MaxHistoricalDays  int
)

// Config holds all configuration for the exchange rate service
type Config struct {
	ServerAddress string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	IdleTimeout   time.Duration
	LogLevel      string
}

// Load reads configuration from environment variables with sensible defaults
// This gets called once during application startup
func Load() *Config {
	// Initialize global config variables from environment
	initializeGlobalConfig()

	return &Config{
		ServerAddress: getEnv("SERVER_ADDRESS", ":"+DefaultServerPort),
		ReadTimeout:   getDurationEnv("READ_TIMEOUT", DefaultAPITimeout),
		WriteTimeout:  getDurationEnv("WRITE_TIMEOUT", DefaultAPITimeout),
		IdleTimeout:   getDurationEnv("IDLE_TIMEOUT", 60*time.Second),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
	}
}

// initializeGlobalConfig loads API-related config from environment variables
// Keeping this separate since these values are used across multiple packages
func initializeGlobalConfig() {
	ExternalAPIBaseURL = getEnv("EXCHANGE_API_BASE_URL", "https://v6.exchangerate-api.com/v6")
	ExchangeRateAPIKey = getEnv("EXCHANGE_API_KEY", "dc07747379a8a53ee8d3243c")
	MaxHistoricalDays = getIntEnv("MAX_HISTORICAL_DAYS", MaxAllowedHistoryDays)

	// Basic validation - we need these to work
	if ExchangeRateAPIKey == "" {
		log.Fatal("EXCHANGE_API_KEY environment variable is required")
	}
}

// getEnv retrieves environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDurationEnv retrieves duration from environment variable or returns default
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getIntEnv retrieves integer environment variable or returns default
// Added this helper since we need it for MaxHistoricalDays config
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// IsSupportedCurrency validates whether a currency code is in our supported list
// We normalize the input to handle different cases and whitespace
func IsSupportedCurrency(code string) bool {
	cleanCode := strings.ToUpper(strings.TrimSpace(code))

	// Quick check for empty input
	if cleanCode == "" {
		return false
	}

	// Linear search is fine for our small currency list
	for _, supportedCode := range SupportedCurrencyList {
		if supportedCode == cleanCode {
			return true
		}
	}
	return false
}

// GetSupportedCurrencies returns a copy of the supported currency list
// Using a function to prevent external modification of our internal slice
func GetSupportedCurrencies() []string {
	currencies := make([]string, len(SupportedCurrencyList))
	copy(currencies, SupportedCurrencyList)
	return currencies
}
