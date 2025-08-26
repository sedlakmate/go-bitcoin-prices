package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"bitcoin-prices/internal/cache"
	"bitcoin-prices/internal/pairs"
)

// KrakenTicker defines the dependency needed from the Kraken client.
type KrakenTicker interface {
	GetLastTradeClosed(ctx context.Context, krakenPairs []string) (map[string]float64, error)
}

// Service orchestrates fetching LTP for supported pairs with per-pair caching.
// It caches by Kraken symbol to reuse across different external pair sets.

type Service struct {
	kraken KrakenTicker
	cache  *cache.TTLCache[string, float64]
}

func New(kr KrakenTicker, ttl time.Duration) *Service {
	return &Service{kraken: kr, cache: cache.New[string, float64](ttl)}
}

// GetLTP returns a map of external pair -> price.
// It fetches missing pairs in batch from Kraken and populates the cache.
func (s *Service) GetLTP(ctx context.Context, extPairs []string) (map[string]float64, error) {
	if len(extPairs) == 0 {
		return nil, errors.New("no pairs provided")
	}
	krSyms := pairs.KrakenSymbols(extPairs)
	missing := make([]string, 0, len(krSyms))
	krPrice := make(map[string]float64, len(krSyms))
	for _, sym := range krSyms {
		if v, ok := s.cache.Get(sym); ok {
			krPrice[sym] = v
		} else {
			missing = append(missing, sym)
		}
	}
	if len(missing) > 0 {
		fresh, err := s.kraken.GetLastTradeClosed(ctx, missing)
		if err != nil {
			return nil, fmt.Errorf("kraken: %w", err)
		}
		for k, v := range fresh {
			krPrice[k] = v
			s.cache.Set(k, v)
		}
	}
	// Map back to external pairs
	ext := pairs.MapKrakenToExternal(extPairs, krPrice)
	return ext, nil
}

// BuildResponse formats the service response payload as required.
// Sorted by pair for deterministic output.
func BuildResponse(extPrices map[string]float64) map[string]any {
	keys := make([]string, 0, len(extPrices))
	for k := range extPrices {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	ltp := make([]map[string]any, 0, len(keys))
	for _, k := range keys {
		ltp = append(ltp, map[string]any{"pair": k, "amount": extPrices[k]})
	}
	return map[string]any{"ltp": ltp}
}

// ParsePairsQuery validates the pairs query string using pairs.NormalizePairs.
func ParsePairsQuery(q string) ([]string, error) {
	ps, err := pairs.NormalizePairs(q)
	if err != nil {
		return nil, err
	}
	return ps, nil
}

// Helper to join pairs for logging
func JoinPairs(ps []string) string { return strings.Join(ps, ",") }
