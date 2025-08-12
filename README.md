# 💱 Exchange Rate Service

A Go backend service for real-time currency exchange rates, following clean architecture principles.
Data is sourced from ExchangeRate-API.com with caching for speed and reliability.

## 📁 Project Structure

```
cmd/server/       → Application entry point
config/           → Configuration & constants
internal/
  handlers/       → HTTP routes & request handling
  services/       → Business logic
  models/         → Domain models
  cache/          → In-memory cache
  client/         → API client for external data
  utils/          → Helper functions
Dockerfile        → Docker configuration
README.md         → Documentation
```

## ✨ Features

- Live Exchange Rates from ExchangeRate-API.com
- Hourly Cache Refresh for fast responses
- Clean Architecture for easy maintenance
- Input Validation with clear error messages
- Health Check Endpoint
- Retry Logic for API requests
- Docker Support for containerized deployment

## 🚀 How to Run

### Prerequisites
- Go 1.21 or higher
- Internet connection for API access

### Running Locally
```bash
# Clone the repository
git clone <repository-url>
cd exchange-rate-service

# Run the service
go run cmd/server/main.go
```

### Using Make
```bash
# Build and run
make run

# Just build
make build

# Run tests
make test

# Build Docker image
make docker-build

# Run with Docker
make docker-run
```

### Testing Endpoints
```bash
# Health check
curl http://localhost:8080/health

# Convert currency
curl "http://localhost:8080/convert?from=USD&to=INR&amount=100"

# Get latest rate
curl "http://localhost:8080/rate/latest?from=USD&to=EUR"

# Get historical rate (within last 90 days)
curl "http://localhost:8080/rate/historical?from=USD&to=EUR&date=2025-08-01"
```

## 🔗 API Integration

- **Base URL**: https://v6.exchangerate-api.com/v6
- **Authentication**: API Key
- **Rate Limit**: 1,500 requests/month (free plan)
- **Supported Currencies**: USD, INR, EUR, JPY, GBP

## 🔍 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Service health check |
| GET | `/convert?from=USD&to=INR&amount=100` | Currency conversion |
| GET | `/rate/latest?from=USD&to=INR` | Latest exchange rate |
| GET | `/rate/historical?from=USD&to=INR&date=YYYY-MM-DD` | Historical exchange rate (last 90 days) |

### Example Responses

**Convert Currency:**
```bash
GET /convert?from=USD&to=INR&amount=100
```
```json
{"amount": 8769.68}
```

**Latest Rate:**
```bash
GET /rate/latest?from=USD&to=EUR
```
```json
{"from":"USD","to":"EUR","rate":0.8606,"date":"latest"}
```

**Historical Rate:**
```bash
GET /rate/historical?from=USD&to=EUR&date=2025-08-01
```
```json
{"from":"USD","to":"EUR","rate":0.8606,"date":"2025-08-01"}
```

## 🏗️ How It Works

1. Service starts and fetches & caches all currency pairs
2. Cache refreshes every hour in the background
3. Requests are served instantly from cache when possible
4. If no cache is available, API data is fetched in real time

## 🐳 Docker

**Build:**
```bash
docker build -t exchange-rate-service .
```

**Run:**
```bash
docker run -p 8080:8080 exchange-rate-service
```

## ⚙️ Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_ADDRESS` | `:8080` | Server listen address |
| `READ_TIMEOUT` | `15s` | HTTP read timeout |
| `WRITE_TIMEOUT` | `15s` | HTTP write timeout |
| `IDLE_TIMEOUT` | `60s` | HTTP idle timeout |


                                                     