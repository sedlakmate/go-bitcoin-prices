# Bitcoin LTP Service

A small Go service that exposes an HTTP API to retrieve the Last Traded Price (LTP) of Bitcoin for these pairs:
- BTC/USD
- BTC/CHF
- BTC/EUR

It fetches data from Kraken's public API, caches results for a short, configurable TTL, and supports concurrent requests.

## API

GET /api/v1/ltp

Query parameters:
- pairs: optional comma-separated list of pairs (e.g., BTC/USD,BTC/EUR).

Response body:
{
  "ltp": [
    { "pair": "BTC/CHF", "amount": 49000.12 },
    { "pair": "BTC/EUR", "amount": 50000.12 },
    { "pair": "BTC/USD", "amount": 52000.12 }
  ]
}

Errors:
- 400 if pairs are invalid
- 504/502 if upstream request fails or times out

## Configuration

Environment variables:
- PORT: HTTP port (default 8080)
- CACHE_TTL: cache TTL in seconds (default 10)
- KRAKEN_BASE_URL: Kraken API base URL (default https://api.kraken.com)
- KRAKEN_RETRIES: Kraken client retries on 429/5xx (default 2)

## Build and run (local)

Pre-requirements:
- Go 1.22+ (for Docker compatibility)

Commands:
```bash
# Run tests
go test ./...

# Build
go build -o bin/bitcoin-prices ./

# Run
PORT=8080 CACHE_TTL=10 ./bin/bitcoin-prices
```

Example requests:
```bash
curl -s http://localhost:8080/api/v1/ltp | jq
curl -s "http://localhost:8080/api/v1/ltp?pairs=BTC/USD,BTC/EUR" | jq
```

## Docker

Build and run with Docker:
```bash
# Build image
docker build -t bitcoin-ltp:latest .

# Run container
docker run --rm -p 8080:8080 -e CACHE_TTL=10 --name bitcoin-ltp bitcoin-ltp:latest

# Call API
curl -s http://localhost:8080/api/v1/ltp | jq
```

## Notes
- Data freshness: The service fetches live data and caches for a short TTL (default 10s), providing accuracy within the last minute.
- Extensibility: Supported pairs and Kraken symbols are centralized in internal/pairs.
- Logging: Basic structured logging using slog for requests and errors.
- Resilience: The Kraken client retries on 429 and 5xx with backoff and uses timeouts.

