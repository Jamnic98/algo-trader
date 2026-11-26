package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"trader-core/internal/bot"
	"trader-core/internal/engine"
	"trader-core/internal/exchange"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func setupAccountRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	m := engine.NewAccountManager(nil, engine.NewLiveAccount("0", "0.001")) // in-memory
	origAccounts, origAccount := accounts, account
	accounts = m
	account = mustDefaultPaper(t, m)
	t.Cleanup(func() { accounts, account = origAccounts, origAccount })

	r := gin.New()
	RegisterAccountRoutes(r.Group("/api/account"))
	return r
}

func mustDefaultPaper(t *testing.T, m *engine.AccountManager) engine.Account {
	t.Helper()
	acc, err := m.Paper("default")
	if err != nil {
		t.Fatalf("default paper: %v", err)
	}
	return acc
}

func doJSON(t *testing.T, r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAccountCreateListGet(t *testing.T) {
	r := setupAccountRouter(t)

	w := doJSON(t, r, http.MethodPost, "/api/account/paper", `{"name":"acc1","balance":"500","fee":"0.002"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}

	// duplicate -> 409
	w = doJSON(t, r, http.MethodPost, "/api/account/paper", `{"name":"acc1"}`)
	if w.Code != http.StatusConflict {
		t.Fatalf("dup: %d %s", w.Code, w.Body.String())
	}

	// bad body -> 400
	w = doJSON(t, r, http.MethodPost, "/api/account/paper", `{"balance":"5"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing name: %d", w.Code)
	}
	w = doJSON(t, r, http.MethodPost, "/api/account/paper", `{"name":"x","balance":"abc"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad balance: %d", w.Code)
	}

	// list contains acc1 + default + live
	w = doJSON(t, r, http.MethodGet, "/api/account/all", "")
	if w.Code != http.StatusOK {
		t.Fatalf("list: %d", w.Code)
	}
	body := w.Body.String()
	for _, want := range []string{`"acc1"`, `"default"`, `"live"`, `"500"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("list body missing %s: %s", want, body)
		}
	}

	// get one
	w = doJSON(t, r, http.MethodGet, "/api/account/acc1", "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"balance":"500"`) {
		t.Fatalf("get acc1: %d %s", w.Code, w.Body.String())
	}

	// unknown -> 404
	w = doJSON(t, r, http.MethodGet, "/api/account/nope", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("unknown: %d", w.Code)
	}
}

func TestAccountTopUpAndReset(t *testing.T) {
	r := setupAccountRouter(t)
	doJSON(t, r, http.MethodPost, "/api/account/paper", `{"name":"acc2","balance":"100"}`)

	w := doJSON(t, r, http.MethodPost, "/api/account/paper/acc2/topup", `{"amount":"25.5"}`)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"balance":"125.5"`) {
		t.Fatalf("topup: %d %s", w.Code, w.Body.String())
	}

	// invalid amount -> 400
	w = doJSON(t, r, http.MethodPost, "/api/account/paper/acc2/topup", `{"amount":"-5"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("neg topup: %d", w.Code)
	}
	// unknown account -> 404
	w = doJSON(t, r, http.MethodPost, "/api/account/paper/ghost/topup", `{"amount":"5"}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("topup unknown: %d", w.Code)
	}

	// reset
	w = doJSON(t, r, http.MethodPost, "/api/account/paper/acc2/reset", `{"balance":"1000"}`)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"balance":"1000"`) {
		t.Fatalf("reset: %d %s", w.Code, w.Body.String())
	}
}

func TestLegacyAccountRoute(t *testing.T) {
	r := setupAccountRouter(t)
	w := doJSON(t, r, http.MethodGet, "/api/account", "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"balance":"10000"`) {
		t.Fatalf("legacy: %d %s", w.Code, w.Body.String())
	}
}

// fakeBalancesBroker is an exchange.Broker with fixed balances for live-sync tests.
type fakeBalancesBroker struct{}

func (fakeBalancesBroker) Name() string { return "fakebal" }
func (fakeBalancesBroker) Klines(symbol, interval string, startMs, endMs int64, limit int) ([]exchange.Kline, error) {
	return nil, nil
}
func (fakeBalancesBroker) StreamKlines(ctx context.Context, symbol, interval string, onKline func(exchange.Kline)) error {
	return nil
}
func (fakeBalancesBroker) Balances() (map[string]decimal.Decimal, error) {
	return map[string]decimal.Decimal{"USD": decimal.NewFromInt(5000), "AAPL": decimal.NewFromInt(3)}, nil
}
func (fakeBalancesBroker) Price(symbol string) (decimal.Decimal, error) { return decimal.NewFromInt(1), nil }
func (fakeBalancesBroker) PlaceMarketOrder(symbol, side string, qty decimal.Decimal, clientOrderID string) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	return decimal.NewFromInt(1), qty, decimal.Zero, nil
}
func (fakeBalancesBroker) Filters(symbol string) (engine.SymbolFilters, error) {
	return engine.SymbolFilters{}, nil
}

// TestLiveAccountPerExchangeAPI covers the per-exchange live ledger endpoints:
// GET /live?exchange=, POST /live/sync?exchange=, and sync-all with no param.
func TestLiveAccountPerExchangeAPI(t *testing.T) {
	r := setupAccountRouter(t)
	exchange.RegisterQuoteAsset("fakebal", "USD")
	origRuntime := runtime
	runtime = &bot.Runtime{Brokers: map[string]exchange.Broker{"fakebal": fakeBalancesBroker{}}}
	t.Cleanup(func() { runtime = origRuntime })

	// default (no param) = binance back-compat
	w := doJSON(t, r, http.MethodGet, "/api/account/live", "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"name":"live:binance"`) {
		t.Fatalf("GET live default: %d %s", w.Code, w.Body.String())
	}

	// registered venue without a ledger yet -> 200, enriched with fresh balances
	w = doJSON(t, r, http.MethodGet, "/api/account/live?exchange=fakebal", "")
	if w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), `"exchange":"fakebal"`) ||
		!strings.Contains(w.Body.String(), `"USD":"5000"`) {
		t.Fatalf("GET live fakebal: %d %s", w.Code, w.Body.String())
	}

	// unknown venue -> 404
	w = doJSON(t, r, http.MethodGet, "/api/account/live?exchange=ghost", "")
	if w.Code != http.StatusNotFound {
		t.Fatalf("GET live ghost: %d %s", w.Code, w.Body.String())
	}

	// sync one venue only
	w = doJSON(t, r, http.MethodPost, "/api/account/live/sync?exchange=fakebal", "")
	if w.Code != http.StatusOK ||
		!strings.Contains(w.Body.String(), `"name":"live:fakebal"`) ||
		!strings.Contains(w.Body.String(), `"balance":"5000"`) ||
		!strings.Contains(w.Body.String(), `"AAPL":"3"`) {
		t.Fatalf("sync fakebal: %d %s", w.Code, w.Body.String())
	}

	// listing now shows both live ledgers
	w = doJSON(t, r, http.MethodGet, "/api/account/all", "")
	if !strings.Contains(w.Body.String(), `"live:binance"`) || !strings.Contains(w.Body.String(), `"live:fakebal"`) {
		t.Fatalf("list missing live ledgers: %s", w.Body.String())
	}

	// get by name works for the per-exchange ledger
	w = doJSON(t, r, http.MethodGet, "/api/account/live:fakebal", "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"balance":"5000"`) {
		t.Fatalf("get live:fakebal: %d %s", w.Code, w.Body.String())
	}

	// sync with no param syncs ALL registered brokers
	w = doJSON(t, r, http.MethodPost, "/api/account/live/sync", "")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"fakebal"`) || !strings.Contains(w.Body.String(), `"live:fakebal"`) {
		t.Fatalf("sync all: %d %s", w.Code, w.Body.String())
	}
}
