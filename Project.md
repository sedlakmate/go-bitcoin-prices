# Task
Develop in Go language a service that will provide an API for retrieval of the Last Traded Price of Bitcoin for the following currency pairs:

1. BTC/USD
2. BTC/CHF
3. BTC/EUR


The request path is:
`/api/v1/ltp`

The response shall constitute JSON of the following structure:
```json
{
  "ltp": [
    {
      "pair": "BTC/CHF",
      "amount": 49000.12
    },
    {
      "pair": "BTC/EUR",
      "amount": 50000.12
    },
    {
      "pair": "BTC/USD",
      "amount": 52000.12
    }
  ]
}

```

# Requirements:

1. The incoming request can done for as for a single pair as well for a list of them
1. You shall provide time accuracy of the data up to the last minute.
1. Code shall be hosted in a remote public repository
1. readme.md includes clear steps to build and run the app
1. Integration tests
1. Dockerized application

# Further considerations:

1. The source of truth for the LTP data shall be Kraken public API. Cross-calculations between like EUR/USD should not be used, especially not from other sources.
1. The service should cache the results for a configurable time period (e.g. 10 seconds) to avoid hitting the Kraken API too often
1. The service should be able to handle concurrent requests
1. The service should log incoming requests and errors
1. The service should validate the incoming request parameters and return appropriate error messages for invalid requests
1. The service should be designed with extensibility in mind, allowing for easy addition of new currency pairs in the future
1. The service should handle potential errors from the Kraken API gracefully, including rate limiting and network issues
1. The service should include unit tests for core functionality

# Docs
The public Kraken API might be used to retrieve the above LTP information
[API Documentation](https://docs.kraken.com/rest/#tag/Spot-Market-Data/operation/getTickerInformation)
(The values of the last traded price is called “last trade closed”)
