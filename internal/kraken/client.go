package kraken

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client fetches ticker info from Kraken public API.
// It supports batching multiple pairs in one request.
// Zero-value is not valid; use NewClient.
type Client struct {
	baseURL string
	http    *http.Client
	retries int
}

func NewClient(baseURL string, httpClient *http.Client, retries int) *Client {
	if baseURL == "" {
		baseURL = "https://api.kraken.com"
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	if retries < 0 {
		retries = 0
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), http: httpClient, retries: retries}
}

// GetLastTradeClosed returns the last trade closed price for each Kraken pair code provided.
// krakenPairs should be Kraken API pair symbols like XBTUSD, XBTEUR, XBTCHF.
func (c *Client) GetLastTradeClosed(ctx context.Context, krakenPairs []string) (map[string]float64, error) {
	if len(krakenPairs) == 0 {
		return map[string]float64{}, nil
	}
	// Deduplicate
	m := make(map[string]struct{}, len(krakenPairs))
	uniq := make([]string, 0, len(krakenPairs))
	for _, p := range krakenPairs {
		if _, ok := m[p]; !ok {
			m[p] = struct{}{}
			uniq = append(uniq, p)
		}
	}

	endpoint := c.baseURL + "/0/public/Ticker"
	q := url.Values{}
	q.Set("pair", strings.Join(uniq, ","))
	u := endpoint + "?" + q.Encode()

	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return nil, err
		}
		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
				lastErr = fmt.Errorf("kraken http %d", resp.StatusCode)
			} else if resp.StatusCode != http.StatusOK {
				b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
				return nil, fmt.Errorf("kraken http %d: %s", resp.StatusCode, string(b))
			} else {
				var tk tickerResponse
				dec := json.NewDecoder(resp.Body)
				if err := dec.Decode(&tk); err != nil {
					return nil, err
				}
				if len(tk.Error) > 0 {
					return nil, errors.New(strings.Join(tk.Error, "; "))
				}
				out := make(map[string]float64, len(tk.Result))
				for pair, data := range tk.Result {
					if len(data.C) >= 1 {
						priceStr := data.C[0]
						f, err := strconv.ParseFloat(priceStr, 64)
						if err != nil {
							return nil, fmt.Errorf("parse price %s for %s: %w", priceStr, pair, err)
						}
						out[pair] = f
					}
				}
				return out, nil
			}
		}
		// Retry with backoff for 429/5xx or network errors
		if attempt < c.retries {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(200*(1<<attempt)) * time.Millisecond):
			}
		}
	}
	return nil, lastErr
}

// Minimal structs matching Kraken response

type tickerResponse struct {
	Error  []string                `json:"error"`
	Result map[string]tickerResult `json:"result"`
}

type tickerResult struct {
	C []string `json:"c"` // last trade closed [price, lot volume]
}
