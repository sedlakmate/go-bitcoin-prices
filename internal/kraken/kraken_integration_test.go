//go:build integration
// +build integration

package kraken

import (
	"context"
	"net/http"
	"testing"
	"time"
)

// This integration test hits the real Kraken API.
// Run with: go test -tags=integration ./internal/kraken -run RealAPI
func TestKrakenClient_GetLastTradeClosed_RealAPI(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	hc := &http.Client{Timeout: 6 * time.Second}
	c := NewClient("https://api.kraken.com", hc, 2)

	// Use modern request symbols; Kraken returns canonicalized keys like XXBTZUSD
	reqPairs := []string{"XBTUSD", "XBTEUR", "XBTCHF"}
	res, err := c.GetLastTradeClosed(ctx, reqPairs)
	if err != nil {
		t.Fatalf("kraken request failed: %v", err)
	}

	// For each requested pair, accept either the simple or canonical response key
	expectedKeys := map[string][]string{
		"XBTUSD": {"XBTUSD", "XXBTZUSD"},
		"XBTEUR": {"XBTEUR", "XXBTZEUR"},
		"XBTCHF": {"XBTCHF", "XXBTZCHF"},
	}

	for _, req := range reqPairs {
		alts := expectedKeys[req]
		found := false
		for _, k := range alts {
			if v, ok := res[k]; ok {
				if v <= 0 {
					t.Fatalf("non-positive price for %s (%s): %v", req, k, v)
				}
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing price for requested pair %s; got keys: %v", req, keysOf(res))
		}
	}
}

func keysOf(m map[string]float64) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}
