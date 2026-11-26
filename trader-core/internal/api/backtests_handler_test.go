package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupBacktestRouter(t *testing.T, lookup func(string) error, backtesterURL string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	orig := strategyLookup
	strategyLookup = lookup
	t.Cleanup(func() { strategyLookup = orig })

	r := gin.New()
	RegisterBacktestRoutes(r.Group("/backtests"), backtesterURL)
	return r
}

func doBacktest(t *testing.T, r *gin.Engine, body string) *httptest.ResponseRecorder {
	t.Helper()
	return doBacktestPath(t, r, "/backtests", body)
}

func doBacktestPath(t *testing.T, r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestBacktestHandlerBadJSON(t *testing.T) {
	r := setupBacktestRouter(t, func(string) error { return nil }, "http://unused")
	w := doBacktest(t, r, `{not json`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", w.Code, w.Body.String())
	}
}

func TestBacktestHandlerMissingFields(t *testing.T) {
	r := setupBacktestRouter(t, func(string) error { return nil }, "http://unused")
	w := doBacktest(t, r, `{"strategy_slug":"simple"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", w.Code, w.Body.String())
	}
}

func TestBacktestHandlerBadInterval(t *testing.T) {
	r := setupBacktestRouter(t, func(string) error { return nil }, "http://unused")
	w := doBacktest(t, r, `{"strategy_slug":"simple","symbol":"BTCUSDT","interval":"7m","start":"2024-01-01","end":"2024-02-01"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "unsupported interval") {
		t.Fatalf("body = %s", w.Body.String())
	}
}

func TestBacktestHandlerBadStart(t *testing.T) {
	r := setupBacktestRouter(t, func(string) error { return nil }, "http://unused")
	w := doBacktest(t, r, `{"strategy_slug":"simple","symbol":"BTCUSDT","interval":"1h","start":"not-a-date","end":"2024-02-01"}`)
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "invalid start") {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestBacktestHandlerEndBeforeStart(t *testing.T) {
	r := setupBacktestRouter(t, func(string) error { return nil }, "http://unused")
	w := doBacktest(t, r, `{"strategy_slug":"simple","symbol":"BTCUSDT","interval":"1h","start":"2024-02-01","end":"2024-01-01"}`)
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "end must be after start") {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestBacktestHandlerStrategyNotFound(t *testing.T) {
	r := setupBacktestRouter(t, func(string) error { return errors.New("record not found") }, "http://unused")
	w := doBacktest(t, r, `{"strategy_slug":"nope","symbol":"BTCUSDT","interval":"1h","start":"2024-01-01","end":"2024-02-01"}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404; body: %s", w.Code, w.Body.String())
	}
}

func TestBacktestHandlerHappyPath(t *testing.T) {
	var gotPayload map[string]any
	backtester := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/backtest" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &gotPayload); err != nil {
			t.Errorf("bad payload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stats":{"num_trades":2}}`))
	}))
	defer backtester.Close()

	lookedUp := ""
	r := setupBacktestRouter(t, func(slug string) error {
		lookedUp = slug
		return nil
	}, backtester.URL)

	w := doBacktest(t, r, `{"strategy_slug":"simple","symbol":"BTCUSDT","interval":"1h","start":"2024-01-01","end":"2024-02-01","quantity":"0.5"}`)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"num_trades":2`) {
		t.Fatalf("response not proxied: %s", w.Body.String())
	}
	if lookedUp != "simple" {
		t.Fatalf("lookup called with %q", lookedUp)
	}
	if gotPayload["strategy_slug"] != "simple" || gotPayload["symbol"] != "BTCUSDT" || gotPayload["interval"] != "1h" {
		t.Fatalf("payload = %v", gotPayload)
	}
	// dates normalized to unix ms
	if gotPayload["start"] != float64(1704067200000) || gotPayload["end"] != float64(1706745600000) {
		t.Fatalf("start/end = %v/%v", gotPayload["start"], gotPayload["end"])
	}
	// defaults applied
	if gotPayload["initial_balance"] != "10000" || gotPayload["fee_rate"] != "0.001" {
		t.Fatalf("defaults = %v/%v", gotPayload["initial_balance"], gotPayload["fee_rate"])
	}
	if gotPayload["quantity"] != "0.5" {
		t.Fatalf("quantity = %v", gotPayload["quantity"])
	}
}

func TestBacktestHandlerBacktesterDown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := server.URL
	server.Close()

	r := setupBacktestRouter(t, func(string) error { return nil }, url)
	w := doBacktest(t, r, `{"strategy_slug":"simple","symbol":"BTCUSDT","interval":"1h","start":"2024-01-01","end":"2024-02-01"}`)
	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502; body: %s", w.Code, w.Body.String())
	}
}

func TestBacktestHandlerForwardsExchange(t *testing.T) {
	var gotPayload map[string]any
	backtester := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &gotPayload); err != nil {
			t.Errorf("bad payload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stats":{"num_trades":1}}`))
	}))
	defer backtester.Close()

	r := setupBacktestRouter(t, func(string) error { return nil }, backtester.URL)
	w := doBacktest(t, r, `{"strategy_slug":"simple","exchange":"alpaca","symbol":"BTC/USD","interval":"1h","start":"2024-01-01","end":"2024-02-01"}`)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	if gotPayload["exchange"] != "alpaca" || gotPayload["symbol"] != "BTC/USD" {
		t.Fatalf("payload = %v", gotPayload)
	}
}

func TestBacktestHandlerDefaultsExchangeToBinance(t *testing.T) {
	var gotPayload map[string]any
	backtester := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotPayload)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stats":{"num_trades":0}}`))
	}))
	defer backtester.Close()

	r := setupBacktestRouter(t, func(string) error { return nil }, backtester.URL)
	w := doBacktest(t, r, `{"strategy_slug":"simple","symbol":"BTCUSDT","interval":"1h","start":"2024-01-01","end":"2024-02-01"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d; body: %s", w.Code, w.Body.String())
	}
	if gotPayload["exchange"] != "binance" {
		t.Fatalf("default exchange = %v, want binance", gotPayload["exchange"])
	}
}

func TestBacktestHandlerMissingSymbol400(t *testing.T) {
	r := setupBacktestRouter(t, func(string) error { return nil }, "http://unused")
	w := doBacktest(t, r, `{"strategy_slug":"simple","interval":"1h","start":"2024-01-01","end":"2024-02-01"}`)
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "symbol or symbols") {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
}

func TestBacktestHandlerForwardsSymbolsAndSlippage(t *testing.T) {
	var gotPayload map[string]any
	backtester := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/backtest" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotPayload)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"stats":{"num_trades":2},"symbols":["BTCUSDT","ETHUSDT"]}`))
	}))
	defer backtester.Close()

	r := setupBacktestRouter(t, func(string) error { return nil }, backtester.URL)
	w := doBacktest(t, r, `{"strategy_slug":"simple","symbols":["BTCUSDT","ETHUSDT"],"interval":"1h","start":"2024-01-01","end":"2024-02-01","slippage_bps":"25"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	symbols, ok := gotPayload["symbols"].([]any)
	if !ok || len(symbols) != 2 || symbols[0] != "BTCUSDT" || symbols[1] != "ETHUSDT" {
		t.Fatalf("symbols = %v", gotPayload["symbols"])
	}
	if gotPayload["slippage_bps"] != "25" {
		t.Fatalf("slippage_bps = %v", gotPayload["slippage_bps"])
	}
	if _, hasSymbol := gotPayload["symbol"]; hasSymbol {
		t.Fatalf("single symbol should be omitted when symbols is set: %v", gotPayload)
	}
	if !strings.Contains(w.Body.String(), `"symbols":["BTCUSDT","ETHUSDT"]`) {
		t.Fatalf("response not proxied: %s", w.Body.String())
	}
}

func TestSweepHandlerForwardsGridToBacktestSweep(t *testing.T) {
	var gotPath string
	var gotPayload map[string]any
	backtester := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotPayload)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":2,"results":[]}`))
	}))
	defer backtester.Close()

	r := setupBacktestRouter(t, func(string) error { return nil }, backtester.URL)
	w := doBacktestPath(t, r, "/backtests/sweep", `{
		"strategy_slug":"sma_cross","symbol":"BTCUSDT","interval":"1h",
		"start":"2024-01-01","end":"2024-02-01","slippage_bps":"10",
		"param_grid":{"fast":[3,5],"slow":[10,20]},"max_combos":50,
		"include_trades":true,"sort_by":"sharpe"
	}`)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	if gotPath != "/backtest/sweep" {
		t.Fatalf("path = %q, want /backtest/sweep", gotPath)
	}
	grid, ok := gotPayload["param_grid"].(map[string]any)
	if !ok || len(grid["fast"].([]any)) != 2 {
		t.Fatalf("param_grid = %v", gotPayload["param_grid"])
	}
	if gotPayload["max_combos"] != float64(50) || gotPayload["include_trades"] != true || gotPayload["sort_by"] != "sharpe" {
		t.Fatalf("payload = %v", gotPayload)
	}
	if gotPayload["slippage_bps"] != "10" || gotPayload["symbol"] != "BTCUSDT" {
		t.Fatalf("payload = %v", gotPayload)
	}
	if !strings.Contains(w.Body.String(), `"count":2`) {
		t.Fatalf("response not proxied: %s", w.Body.String())
	}
}

func TestSweepHandlerValidation(t *testing.T) {
	r := setupBacktestRouter(t, func(string) error { return nil }, "http://unused")
	// missing param_grid -> 400 (binding required)
	w := doBacktestPath(t, r, "/backtests/sweep", `{"strategy_slug":"sma_cross","symbol":"BTCUSDT","interval":"1h","start":"2024-01-01","end":"2024-02-01"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing grid: status = %d, body = %s", w.Code, w.Body.String())
	}
	// missing symbol -> 400
	w = doBacktestPath(t, r, "/backtests/sweep", `{"strategy_slug":"sma_cross","interval":"1h","start":"2024-01-01","end":"2024-02-01","param_grid":{"fast":[3]}}`)
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "symbol or symbols") {
		t.Fatalf("missing symbol: status = %d, body = %s", w.Code, w.Body.String())
	}
	// unknown strategy -> 404
	r404 := setupBacktestRouter(t, func(string) error { return errors.New("record not found") }, "http://unused")
	w = doBacktestPath(t, r404, "/backtests/sweep", `{"strategy_slug":"nope","symbol":"BTCUSDT","interval":"1h","start":"2024-01-01","end":"2024-02-01","param_grid":{"fast":[3]}}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("unknown strategy: status = %d, body = %s", w.Code, w.Body.String())
	}
}
