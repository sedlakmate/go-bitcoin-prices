package pairs

import (
	"fmt"
	"sort"
	"strings"
)

// Supported external pairs and Kraken canonical pair symbols.
// We'll use Kraken classic pair codes so response keys match exactly.
// BTC maps to Kraken's XBT. Classic codes have X/Z prefixes.
var (
	Supported = []string{"BTC/USD", "BTC/EUR", "BTC/CHF"}

	toKraken = map[string]string{
		"BTC/USD": "XXBTZUSD",
		"BTC/EUR": "XXBTZEUR",
		"BTC/CHF": "XXBTZCHF",
	}
)

// NormalizePairs parses a comma-separated list from query and validates.
// Returns sorted unique external pair strings.
func NormalizePairs(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		// default: all
		out := make([]string, len(Supported))
		copy(out, Supported)
		return out, nil
	}
	split := strings.Split(raw, ",")
	set := make(map[string]struct{}, len(split))
	for _, item := range split {
		p := strings.ToUpper(strings.TrimSpace(item))
		if p == "" {
			continue
		}
		if _, ok := toKraken[p]; !ok {
			return nil, fmt.Errorf("unsupported pair: %s", p)
		}
		set[p] = struct{}{}
	}
	if len(set) == 0 {
		return nil, fmt.Errorf("no valid pairs provided")
	}
	out := make([]string, 0, len(set))
	for p := range set {
		out = append(out, p)
	}
	sort.Strings(out)
	return out, nil
}

// KrakenSymbols returns Kraken pair codes for the provided external pairs.
func KrakenSymbols(extPairs []string) []string {
	out := make([]string, 0, len(extPairs))
	for _, p := range extPairs {
		if sym, ok := toKraken[p]; ok {
			out = append(out, sym)
		}
	}
	return out
}

// MapKrakenToExternal builds a mapping from external pair to the price from the Kraken result map.
func MapKrakenToExternal(extPairs []string, krakenPrices map[string]float64) map[string]float64 {
	out := make(map[string]float64, len(extPairs))
	for _, p := range extPairs {
		if sym, ok := toKraken[p]; ok {
			if price, ok2 := krakenPrices[sym]; ok2 {
				out[p] = price
			}
		}
	}
	return out
}
