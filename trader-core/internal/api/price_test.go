package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"trader-core/internal/bot"
	"trader-core/internal/exchange"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type fakePriceBroker struct{ price decimal.Decimal }

func (f fakePriceBroker) Name() string { return "fakepx" }
func (f fakePriceBroker) Klines(symbol, interval string, startMs, endMs int64, limit int) ([]exchange.Kline, error) {
	return nil, nil
}
func (f fakePriceBroker) StreamKlines(ctx context.Context, symbol, interval string, onKline func(exchange.Kline)) error {
	return nil
}
func (f fakePriceBroker) Balances() (map[string]decimal.Decimal, error) { return nil, nil }
func (f fakePriceBroker) Price(symbol string) (decimal.Decimal, error)    { return f.price, nil }
func (f fakePriceBroker) PlaceMarketOrder(symbol string, side string, qty decimal.Decimal) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	return decimal.Zero, decimal.Zero, decimal.Zero, nil
}

func setupPriceRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	orig := runtime
	runtime = &bot.Runtime{Brokers: map[string]exchange.Broker{"fakepx": fakePriceBroker{decimal.NewFromInt(42)}}}
	t.Cleanup(func() { runtime = orig })
	r := gin.New()
	RegisterPriceRoutes(r.Group("/api/price"))
	return r
}

func TestPriceHandlerViaBroker(t *testing.T) {
	r := setupPriceRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/price?symbol=AAPL&exchange=fakepx", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["price"] != "42" || body["exchange"] != "fakepx" {
		t.Fatalf("body = %v", body)
	}
}

func TestPriceHandlerDefaultsToBinance(t *testing.T) {
	r := setupPriceRouter(t)
	// "binance" not in the fake registry -> 400 from BrokerFor, proving the default is binance
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/price?symbol=BTCUSDT", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 (binance not registered in this test); body = %s", w.Code, w.Body.String())
	}
}

func TestPriceHandlerMissingSymbol(t *testing.T) {
	r := setupPriceRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/price?exchange=fakepx", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestPriceHandlerUnknownExchange(t *testing.T) {
	r := setupPriceRouter(t)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/price?symbol=X&exchange=nope", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body = %s", w.Code, w.Body.String())
	}
}
