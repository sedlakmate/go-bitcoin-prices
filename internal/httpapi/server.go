package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"bitcoin-prices/internal/kraken"
	"bitcoin-prices/internal/service"
)

type Server struct {
	log     *slog.Logger
	server  *http.Server
	service *service.Service
}

// NewServer builds an HTTP server bound to addr.
// Env:
// - CACHE_TTL (seconds, default 10)
// - KRAKEN_BASE_URL (default https://api.kraken.com)
// - KRAKEN_RETRIES (default 2)
func NewServer(addr string) *Server {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Config
	ttl := parseEnvInt("CACHE_TTL", 10)
	krBase := getenv("KRAKEN_BASE_URL", "https://api.kraken.com")
	retries := parseEnvInt("KRAKEN_RETRIES", 2)

	kc := kraken.NewClient(krBase, &http.Client{Timeout: 5 * time.Second}, retries)
	svc := service.New(kc, time.Duration(ttl)*time.Second)

	mux := NewHandler(logger, svc)

	hs := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	return &Server{log: logger, server: hs, service: svc}
}

// NewHandler builds the HTTP handler (mux) for the API using provided logger and service.
func NewHandler(logger *slog.Logger, svc *service.Service) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	mux.Handle("/api/v1/ltp", withLogging(logger, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		pairsQ := r.URL.Query().Get("pairs")
		ps, err := service.ParsePairsQuery(pairsQ)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
		defer cancel()

		prices, err := svc.GetLTP(ctx, ps)
		if err != nil {
			code := http.StatusBadGateway
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				code = http.StatusGatewayTimeout
			}
			writeJSON(w, code, map[string]any{"error": "failed to fetch prices"})
			logger.Error("ltp fetch failed", "err", err, "pairs", service.JoinPairs(ps))
			return
		}
		payload := service.BuildResponse(prices)
		writeJSON(w, http.StatusOK, payload)
	})))
	return mux
}

func (s *Server) Start() error {
	s.log.Info("Starting HTTP server", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("Shutting down HTTP server")
	return s.server.Shutdown(ctx)
}

// withLogging is a middleware that logs requests using the provided logger.
func withLogging(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &respWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(rw, r)
		lat := time.Since(start)
		log.Info("http", "method", r.Method, "path", r.URL.Path, "query", r.URL.RawQuery, "ip", clientIP(r), "status", rw.status, "dur_ms", lat.Milliseconds())
	})
}

type respWriter struct {
	http.ResponseWriter
	status int
}

func (w *respWriter) WriteHeader(code int) { w.status = code; w.ResponseWriter.WriteHeader(code) }

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(true)
	_ = enc.Encode(v)
}

func clientIP(r *http.Request) string {
	// X-Forwarded-For first if present
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return i
}
