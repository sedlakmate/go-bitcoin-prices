# Bitcoin Last Traded Price service

[![CI](https://github.com/sedlakmate/go-bitcoin-prices/actions/workflows/ci.yml/badge.svg)](https://github.com/OWNER/REPO/actions/workflows/ci.yml)
[![Integration (Real Kraken)](https://github.com/sedlakmate/go-bitcoin-prices/actions/workflows/integration.yml/badge.svg)](https://github.com/OWNER/REPO/actions/workflows/integration.yml)

A small Go service that exposes an HTTP API to retrieve the Last Traded Price (LTP) of Bitcoin for these pairs:
- BTC/USD
- BTC/CHF
- BTC/EUR

It fetches data from Kraken's public API, caches results for a short, configurable TTL, and supports concurrent requests.

## Considerations

- Written in Go 1.22+ for compatibility with modern Docker images.
- Uses TTL caching to reduce upstream calls and improve performance.
- Basic structured logging with slog.
- Resilient Kraken client with retries and timeouts.
- Unit tests for core logic and handlers.
- Optional integration test against the real Kraken API (caching disabled).
- GitHub Actions CI for linting, testing, and building.

## API

### Health check

`GET /api/health`

Response body:
```
"ok"
```

Example:
`curl -s "http://localhost:8080/api/health"`

### LTP

`GET /api/v1/ltp`

Query parameters:
- `pairs`: comma-separated list of pairs. Possible values: BTC/USD BTC/EUR BTC/CHF


Example:
`curl -s "http://localhost:8080/api/v1/ltp?pairs=BTC/USD,BTC/EUR" | jq `

Response body:
```
{
  "ltp": [
    { "pair": "BTC/CHF", "amount": 49000.12 },
    { "pair": "BTC/EUR", "amount": 50000.12 },
    { "pair": "BTC/USD", "amount": 52000.12 }
  ]
}
```

Errors:
- 400 if pairs are invalid
- 504/502 if upstream request fails or times out

## Configuration

Environment variables:
- PORT: HTTP port (default 8080)
- CACHE_TTL: cache TTL in seconds (default 10)
- KRAKEN_BASE_URL: Kraken API base URL (default https://api.kraken.com)
- KRAKEN_RETRIES: Kraken client retries on 429/5xx (default 2)

## Build and run 

### Locally

Pre-requirements:
- Go 1.22+ (for Docker compatibility)

Commands:
```bash
# Run tests (unit + handler)
go test ./...

# Build
go build -o bin/bitcoin-prices ./

# Run
PORT=8080 CACHE_TTL=10 ./bin/bitcoin-prices
```

Example requests:
```bash
curl -s "http://localhost:8080/api/v1/ltp?pairs=BTC/USD,BTC/EUR" | jq
```

### In docker

Build and run with Docker:
```bash
# Build image
docker build -t bitcoin-ltp:latest .

# Run container
docker run --rm -p 8080:8080 -e CACHE_TTL=10 --name bitcoin-ltp bitcoin-ltp:latest

# Call API
curl -s http://localhost:8080/api/v1/ltp?pairs=BTC/USD,BTC/EUR | jq
```

## Integration tests (real Kraken API)

There is an opt-in integration test that calls the real Kraken API. It’s excluded by default and requires network access.

Run it explicitly with build tags and disable Go test result caching (-count=1):
```bash
# Only run the Kraken integration test, forcing re-execution each time
go test -tags=integration -count=1 ./internal/kraken -run RealAPI -v
```
Notes:
- May be flaky due to network or Kraken rate limiting; re-run if needed.
- Uses pairs XBTUSD/XBTEUR and accepts canonical response keys.


## Notes
- Data freshness: The service fetches live data and caches for a short TTL (default 10s), providing accuracy within the last minute.
- Extensibility: Supported pairs and Kraken symbols are centralized in internal/pairs.
- Logging: Basic structured logging using slog for requests and errors.
- Resilience: The Kraken client retries on 429 and 5xx with backoff and uses timeouts.
