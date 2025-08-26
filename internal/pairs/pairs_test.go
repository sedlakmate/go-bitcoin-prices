package pairs

import "testing"

func TestNormalizePairs_DefaultAll(t *testing.T) {
	ps, err := NormalizePairs("")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(ps) != len(Supported) {
		t.Fatalf("expected %d pairs, got %d", len(Supported), len(ps))
	}
}

func TestNormalizePairs_Specific(t *testing.T) {
	ps, err := NormalizePairs("btc/usd, BTC/EUR ,btc/chf")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(ps) != 3 {
		t.Fatalf("expected 3, got %d", len(ps))
	}
}

func TestNormalizePairs_Invalid(t *testing.T) {
	if _, err := NormalizePairs("ETH/USD"); err == nil {
		t.Fatalf("expected error for invalid pair")
	}
}
