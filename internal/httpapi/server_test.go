package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bitcoin-prices/internal/service"
)

type mockKraken struct {
	resp map[string]float64
	err  error
}

func (m *mockKraken) GetLastTradeClosed(ctx context.Context, pairs []string) (map[string]float64, error) {
	// Return subset that matches requested pairs
	out := make(map[string]float64)
	for _, p := range pairs {
		if v, ok := m.resp[p]; ok {
			out[p] = v
		}
	}
	return out, m.err
}

func newTestHandler() http.Handler {
	mk := &mockKraken{resp: map[string]float64{
		"XXBTZUSD": 52000.12,
		"XXBTZEUR": 50000.12,
		"XXBTZCHF": 49000.12,
	}}
	svc := service.New(mk, time.Minute)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewHandler(logger, svc)
}

func TestLTP_DefaultAll(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest("GET", "/api/v1/ltp", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	ltp, ok := body["ltp"].([]any)
	if !ok || len(ltp) != 3 {
		t.Fatalf("expected 3 items, got %#v", body["ltp"])
	}
}

func TestLTP_SpecificPairs(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest("GET", "/api/v1/ltp?pairs=BTC/USD,BTC/EUR", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var body struct {
		LTP []map[string]any `json:"ltp"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if len(body.LTP) != 2 {
		t.Fatalf("expected 2 items, got %d", len(body.LTP))
	}
}

func TestLTP_InvalidPair(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest("GET", "/api/v1/ltp?pairs=ETH/USD", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 400 {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestLTP_MethodNotAllowed(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest("POST", "/api/v1/ltp", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 405 {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
