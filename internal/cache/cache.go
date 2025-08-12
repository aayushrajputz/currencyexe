package cache

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"exchange-rate-service/config"
)

// Cache defines the interface for caching operations
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// cache for exchange rates with bg refresh
type ExchangeRateCache struct {
	rateMutex sync.RWMutex
	rateData  map[string]rateEntry

	// api client for fetching rates
	exchangeAPIClient ExchangeRateAPIClient
	shutdownChannel   chan struct{}
	backgroundWorkers sync.WaitGroup
}

// rateEntry holds a single exchange rate with its timestamp

type rateEntry struct {
	exchangeRate float64
	lastUpdated  time.Time
}

// ExchangeRateAPIClient defines what we need from our API client

type ExchangeRateAPIClient interface {
	GetRate(fromCurrency, toCurrency, dateStr string) (float64, error)
}

// NewExchangeRateCache creates a new cache instance with the provided API client
func NewExchangeRateCache(apiClient ExchangeRateAPIClient) *ExchangeRateCache {
	return &ExchangeRateCache{
		rateData:          make(map[string]rateEntry),
		exchangeAPIClient: apiClient,
		shutdownChannel:   make(chan struct{}),
	}
}

// GetRate retrieves a cached exchange rate if it exists

func (cache *ExchangeRateCache) GetRate(fromCurrency, toCurrency string) (float64, bool) {
	cacheKey := buildRateKey(fromCurrency, toCurrency)

	cache.rateMutex.RLock()
	entry, found := cache.rateData[cacheKey]
	cache.rateMutex.RUnlock()

	if !found {
		return 0, false
	}

	return entry.exchangeRate, true
}

// SetRate stores an exchange rate in the cache with current timestamp
func (cache *ExchangeRateCache) SetRate(fromCurrency, toCurrency string, rate float64) {
	cacheKey := buildRateKey(fromCurrency, toCurrency)

	cache.rateMutex.Lock()
	cache.rateData[cacheKey] = rateEntry{
		exchangeRate: rate,
		lastUpdated:  time.Now(),
	}
	cache.rateMutex.Unlock()
}

// This runs in a separate goroutine to avoid blocking the main application
func (cache *ExchangeRateCache) StartHourlyRefresh() {
	cache.backgroundWorkers.Add(1)
	go cache.refreshLoop()
}

// Stop gracefully shuts down the refresh process and waits for completion
func (cache *ExchangeRateCache) Stop() {
	close(cache.shutdownChannel)
	cache.backgroundWorkers.Wait()
}

// refreshLoop runs the hourly refresh cycle in the background
// This is the main worker goroutine that keeps our exchange rates current
func (cache *ExchangeRateCache) refreshLoop() {
	defer cache.backgroundWorkers.Done()

	// Use the configured refresh interval from our constants
	refreshTicker := time.NewTicker(config.CacheRefreshInterval)
	defer refreshTicker.Stop()

	// Do an initial refresh right away when we start up
	cache.refreshAllRates()

	for {
		select {
		case <-refreshTicker.C:
			cache.refreshAllRates()
		case <-cache.shutdownChannel:
			return
		}
	}
}

// This is called periodically by the background refresh goroutine
func (cache *ExchangeRateCache) refreshAllRates() {
	supportedCurrencies := config.GetSupportedCurrencies()
	successfulUpdates := 0
	totalPairs := 0
	failedPairs := make([]string, 0)

	log.Printf("Starting exchange rate refresh for %d currencies", len(supportedCurrencies))

	// Iterate through all currency pair combinations
	for i, fromCurrency := range supportedCurrencies {
		for j, toCurrency := range supportedCurrencies {
			// Skip same-currency pairs (USD->USD doesn't make sense)
			if i == j {
				continue
			}

			totalPairs++
			pairIdentifier := fmt.Sprintf("%s-%s", fromCurrency, toCurrency)

			// Fetch the latest rate from our API client
			exchangeRate, err := cache.exchangeAPIClient.GetRate(fromCurrency, toCurrency, "")
			if err != nil {
				log.Printf("Failed to fetch rate %s: %v", pairIdentifier, err)
				failedPairs = append(failedPairs, pairIdentifier)
				continue
			}

			// Store the successful rate in our cache
			cache.SetRate(fromCurrency, toCurrency, exchangeRate)
			successfulUpdates++

			// Log the first few successful fetches for debugging
			if successfulUpdates <= 3 {
				log.Printf("Successfully fetched rate %s: %.6f", pairIdentifier, exchangeRate)
			}
		}
	}

	// Report the final results of this refresh cycle
	if len(failedPairs) > 0 {
		log.Printf("Exchange rate refresh completed: %d/%d pairs updated successfully. Failed pairs: %v",
			successfulUpdates, totalPairs, failedPairs)
	} else {
		log.Printf("Exchange rate refresh completed: %d/%d pairs updated successfully", successfulUpdates, totalPairs)
	}

}

// buildRateKey creates a cache key for currency pair
func buildRateKey(from, to string) string {
	fromClean := strings.ToUpper(strings.TrimSpace(from))
	toClean := strings.ToUpper(strings.TrimSpace(to))
	return fmt.Sprintf("%s-%s", fromClean, toClean)
}

// GetCacheStats returns statistics about cached rates
func (cache *ExchangeRateCache) GetCacheStats() map[string]interface{} {
	cache.rateMutex.RLock()
	defer cache.rateMutex.RUnlock()

	stats := make(map[string]interface{})
	stats["total_pairs"] = len(cache.rateData)

	if len(cache.rateData) > 0 {
		var oldestUpdate time.Time
		var newestUpdate time.Time

		isFirstEntry := true
		for _, entry := range cache.rateData {
			if isFirstEntry {
				oldestUpdate = entry.lastUpdated
				newestUpdate = entry.lastUpdated
				isFirstEntry = false
				continue
			}

			if entry.lastUpdated.Before(oldestUpdate) {
				oldestUpdate = entry.lastUpdated
			}
			if entry.lastUpdated.After(newestUpdate) {
				newestUpdate = entry.lastUpdated
			}
		}

		stats["oldest_update"] = oldestUpdate
		stats["newest_update"] = newestUpdate
	}

	return stats
}
