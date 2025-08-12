package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"exchange-rate-service/config"
	"exchange-rate-service/internal/cache"
	"exchange-rate-service/internal/client"
	"exchange-rate-service/internal/handlers"
	"exchange-rate-service/internal/services"

	"github.com/gorilla/mux"
)

func main() {
	log.Println("Starting Exchange Rate Service...")

	// load config
	cfg := config.Load()
	log.Printf("Server will listen on %s", cfg.ServerAddress)

	// setup api client
	apiClient := client.NewRateClient()
	log.Println("Exchange rate API client initialized")

	// cache setup - auto refresh every hour
	rateCache := cache.NewExchangeRateCache(apiClient)
	rateCache.StartHourlyRefresh()
	defer rateCache.Stop()
	log.Println("Background rate refresh started")

	// services
	healthSvc := services.NewHealthService()
	exchangeSvc := services.NewCurrencyExchangeService(rateCache, apiClient)

	// handlers
	healthHandler := handlers.NewHealthHandler(healthSvc)
	exchangeHandler := handlers.NewExchangeHandler(exchangeSvc)

	// setup routes
	router := mux.NewRouter()
	setupRoutes(router, healthHandler, exchangeHandler)

	// http server config
	srv := &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// start server
	go func() {
		log.Printf("Starting exchange rate service on %s", cfg.ServerAddress)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server startup failed: %v", err)
		}
	}()

	// wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func setupRoutes(router *mux.Router, healthHandler *handlers.HealthHandler, exchangeHandler *handlers.ExchangeHandler) {
	// health endpoint
	router.HandleFunc("/health", healthHandler.CheckHealth).Methods("GET")

	// exchange endpoints
	router.HandleFunc("/convert", exchangeHandler.Convert).Methods("GET")
	router.HandleFunc("/rate/latest", exchangeHandler.GetLatestRate).Methods("GET")
	router.HandleFunc("/rate/historical", exchangeHandler.GetHistoricalRate).Methods("GET")

	// middleware
	router.Use(loggingMiddleware)
	router.Use(recoveryMiddleware)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
