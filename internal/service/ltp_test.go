package service

import (
	"context"
	"testing"
	"time"
)

type mockKraken struct {
	calls int
	resp  map[string]float64
	err   error
}

func (m *mockKraken) GetLastTradeClosed(ctx context.Context, krakenPairs []string) (map[string]float64, error) {
	m.calls++
	// return only requested keys that we have
	out := make(map[string]float64)
	for _, k := range krakenPairs {
		if v, ok := m.resp[k]; ok {
			out[k] = v
		}
	}
	return out, m.err
}

func TestService_GetLTP_CachesResults(t *testing.T) {
	mk := &mockKraken{resp: map[string]float64{
		"XXBTZUSD": 52000.12,
		"XXBTZEUR": 50000.12,
	}}
	s := New(mk, 500*time.Millisecond)
	ctx := context.Background()
	pairs := []string{"BTC/USD", "BTC/EUR"}

	res1, err := s.GetLTP(ctx, pairs)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(res1) != 2 {
		t.Fatalf("expected 2 results, got %d", len(res1))
	}
	if mk.calls != 1 {
		t.Fatalf("expected 1 kraken call, got %d", mk.calls)
	}

	res2, err := s.GetLTP(ctx, pairs)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(res2) != 2 {
		t.Fatalf("expected 2 results, got %d", len(res2))
	}
	if mk.calls != 1 {
		t.Fatalf("expected cache hit (no new kraken call), got %d", mk.calls)
	}
}
