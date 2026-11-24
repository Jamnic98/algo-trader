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
	r.POST("/backtests", func(c *gin.Context) { runBacktestHandler(c, backtesterURL) })
	return r
}

func doBacktest(t *testing.T, r *gin.Engine, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/backtests", strings.NewReader(body))
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
