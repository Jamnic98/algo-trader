package alpaca

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trader-core/internal/exchange"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

func testBroker(server *httptest.Server) *Broker {
	return New(exchange.Config{
		APIKey:  "k",
		Secret:  "s",
		BaseURL: server.URL,
		DataURL: server.URL,
		WSURL:   "ws" + strings.TrimPrefix(server.URL, "http") + "/ws",
		Paper:   true,
	})
}

func TestRegisteredAsAlpaca(t *testing.T) {
	b, err := exchange.New("alpaca", exchange.Config{APIKey: "k", Secret: "s", Paper: true})
	if err != nil {
		t.Fatalf("New(alpaca): %v", err)
	}
	if b.Name() != "alpaca" {
		t.Fatalf("name = %q", b.Name())
	}
	// defaults: paper trading URL + data URL
	ab := b.(*Broker)
	if ab.cfg.BaseURL != defaultTradingURLPaper {
		t.Fatalf("paper default base = %q", ab.cfg.BaseURL)
	}
}

func TestTimeframeMapping(t *testing.T) {
	cases := map[string]string{"1m": "1Min", "5m": "5Min", "1h": "1Hour", "1d": "1Day", "1w": "1Week"}
	for in, want := range cases {
		got, err := timeframe(in)
		if err != nil || got != want {
			t.Errorf("timeframe(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	if _, err := timeframe("7m"); err == nil {
		t.Error("expected error for 7m")
	}
}

func TestSymbolHelpers(t *testing.T) {
	if got := SymbolFor("btc", "usd", true); got != "BTC/USD" {
		t.Errorf("SymbolFor crypto = %q", got)
	}
	if got := SymbolFor("aapl", "", false); got != "AAPL" {
		t.Errorf("SymbolFor stock = %q", got)
	}
	if !IsCryptoSymbol("BTC/USD") || IsCryptoSymbol("AAPL") {
		t.Error("IsCryptoSymbol wrong")
	}
}

func TestKlinesStocksPaginates(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v2/stocks/AAPL/bars") {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("APCA-API-KEY-ID") != "k" {
			t.Errorf("missing auth header")
		}
		q := r.URL.Query()
		if q.Get("timeframe") != "1Hour" || q.Get("feed") != "iex" {
			t.Errorf("params = %v", q)
		}
		calls++
		if q.Get("page_token") == "" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"bars": []map[string]any{
					{"t": "2024-01-01T14:00:00Z", "o": 100, "h": 101, "l": 99, "c": 100.5, "v": 1000},
					{"t": "2024-01-01T15:00:00Z", "o": 100.5, "h": 102, "l": 100, "c": 101.5, "v": 1200},
				},
				"next_page_token": "page2",
			})
		} else {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"bars": []map[string]any{
					{"t": "2024-01-01T16:00:00Z", "o": 101.5, "h": 103, "l": 101, "c": 102.5, "v": 900},
				},
			})
		}
	}))
	defer server.Close()

	b := testBroker(server)
	kl, err := b.Klines("AAPL", "1h", 0, 0, 10)
	if err != nil {
		t.Fatalf("Klines: %v", err)
	}
	if len(kl) != 3 || calls != 2 {
		t.Fatalf("klines = %d, calls = %d", len(kl), calls)
	}
	if !kl[0].Close.Equal(decimal.NewFromFloat(100.5)) || kl[2].OpenTime != time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC).UnixMilli() {
		t.Fatalf("klines = %+v", kl)
	}
}

func TestKlinesCryptoPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.EscapedPath(), "/v1beta3/crypto/us/BTC%2FUSD/bars") {
			t.Errorf("unexpected path %s", r.URL.EscapedPath())
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"bars": []map[string]any{
				{"t": "2024-01-01T00:00:00Z", "o": 42000, "h": 42100, "l": 41900, "c": 42050, "v": 12.5},
			},
		})
	}))
	defer server.Close()

	b := testBroker(server)
	kl, err := b.Klines("BTC/USD", "1h", 0, 0, 10)
	if err != nil || len(kl) != 1 {
		t.Fatalf("Klines crypto: %v %v", kl, err)
	}
	if !kl[0].Close.Equal(decimal.NewFromFloat(42050)) {
		t.Fatalf("close = %s", kl[0].Close)
	}
}

func TestBalances(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v2/account":
			_ = json.NewEncoder(w).Encode(map[string]string{"cash": "5000.25"})
		case "/v2/positions":
			_ = json.NewEncoder(w).Encode([]map[string]string{
				{"symbol": "AAPL", "qty": "10"},
				{"symbol": "TSLA", "qty": "0"},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	b := testBroker(server)
	balances, err := b.Balances()
	if err != nil {
		t.Fatalf("Balances: %v", err)
	}
	if !balances["USD"].Equal(decimal.NewFromFloat(5000.25)) {
		t.Errorf("USD = %v", balances["USD"])
	}
	if !balances["AAPL"].Equal(decimal.NewFromInt(10)) {
		t.Errorf("AAPL = %v", balances["AAPL"])
	}
	if _, ok := balances["TSLA"]; ok {
		t.Error("zero position should be omitted")
	}
}

func TestBalancesRequireCreds(t *testing.T) {
	b := New(exchange.Config{})
	if _, err := b.Balances(); err == nil {
		t.Fatal("expected creds error")
	}
	if _, _, _, err := b.PlaceMarketOrder("AAPL", "BUY", decimal.NewFromInt(1), "test-order-id"); err == nil {
		t.Fatal("expected creds error")
	}
}

func TestPlaceMarketOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/orders" || r.Method != http.MethodPost {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["symbol"] != "AAPL" || body["side"] != "buy" || body["type"] != "market" || body["qty"] != "2" {
			t.Errorf("body = %v", body)
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":               "ord-1",
			"status":           "filled",
			"filled_qty":       "2",
			"filled_avg_price": "187.42",
		})
	}))
	defer server.Close()

	b := testBroker(server)
	avg, qty, fee, err := b.PlaceMarketOrder("aapl", "BUY", decimal.NewFromInt(2), "test-order-id")
	if err != nil {
		t.Fatalf("PlaceMarketOrder: %v", err)
	}
	if !avg.Equal(decimal.NewFromFloat(187.42)) || !qty.Equal(decimal.NewFromInt(2)) || !fee.IsZero() {
		t.Fatalf("avg=%s qty=%s fee=%s", avg, qty, fee)
	}
}

func TestPlaceMarketOrderRejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":            "ord-2",
			"status":        "rejected",
			"reject_reason": "insufficient buying power",
		})
	}))
	defer server.Close()

	b := testBroker(server)
	_, _, _, err := b.PlaceMarketOrder("AAPL", "BUY", decimal.NewFromInt(1000000), "test-order-id")
	if err == nil || !strings.Contains(err.Error(), "insufficient buying power") {
		t.Fatalf("expected reject error, got %v", err)
	}
}

func TestStreamKlinesEmitsClosedBars(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		// read auth + subscribe
		var auth, sub map[string]any
		_ = conn.ReadJSON(&auth)
		_ = conn.ReadJSON(&sub)
		if auth["action"] != "auth" || sub["action"] != "subscribe" {
			t.Errorf("auth=%v sub=%v", auth, sub)
		}
		_ = conn.WriteJSON([]map[string]any{{"T": "success", "msg": "authenticated"}})
		// two bars with different timestamps -> first one is "closed" when second arrives
		_ = conn.WriteJSON([]map[string]any{
			{"T": "b", "S": "AAPL", "t": "2024-01-01T14:00:00Z", "o": 100, "h": 101, "l": 99, "c": 100.5, "v": 10},
			{"T": "b", "S": "AAPL", "t": "2024-01-01T15:00:00Z", "o": 100.5, "h": 102, "l": 100, "c": 101, "v": 12},
		})
		time.Sleep(300 * time.Millisecond)
	}))
	defer server.Close()

	b := testBroker(server)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var got []exchange.Kline
	// server closes after sending; either ctx timeout or EOF is fine — we only care about emitted bars
	_ = b.StreamKlines(ctx, "AAPL", "1h", func(k exchange.Kline) { got = append(got, k) })
	if len(got) != 1 || !got[0].Close.Equal(decimal.NewFromFloat(100.5)) {
		t.Fatalf("emitted klines = %+v", got)
	}
}


func TestPlaceMarketOrderPollsUntilFilled(t *testing.T) {
	polls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v2/orders":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":         "ord-9",
				"status":     "accepted",
				"filled_qty": "0",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/v2/orders/ord-9":
			polls++
			if polls < 2 {
				_ = json.NewEncoder(w).Encode(map[string]any{"id": "ord-9", "status": "accepted", "filled_qty": "0"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id":               "ord-9",
				"status":           "filled",
				"filled_qty":       "2",
				"filled_avg_price": "187.42",
			})
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	b := testBroker(server)
	avg, qty, fee, err := b.placeMarketOrder("AAPL", "BUY", decimal.NewFromInt(2), "test-order-id", 5*time.Millisecond, 2*time.Second)
	if err != nil {
		t.Fatalf("placeMarketOrder: %v", err)
	}
	if polls < 2 {
		t.Fatalf("expected polling, got %d polls", polls)
	}
	if !avg.Equal(decimal.NewFromFloat(187.42)) || !qty.Equal(decimal.NewFromInt(2)) || !fee.IsZero() {
		t.Fatalf("avg=%s qty=%s fee=%s", avg, qty, fee)
	}
}

func TestPlaceMarketOrderCanceledAfterAccept(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v2/orders" {
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": "ord-10", "status": "accepted", "filled_qty": "0"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "ord-10", "status": "canceled", "filled_qty": "0"})
	}))
	defer server.Close()

	b := testBroker(server)
	_, _, _, err := b.placeMarketOrder("AAPL", "BUY", decimal.NewFromInt(1), "test-order-id", 5*time.Millisecond, time.Second)
	if err == nil || !strings.Contains(err.Error(), "not filled") {
		t.Fatalf("expected not-filled error, got %v", err)
	}
}

func TestPriceStocks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/stocks/AAPL/trades/latest" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if r.URL.Query().Get("feed") != "iex" {
			t.Errorf("feed = %q", r.URL.Query().Get("feed"))
		}
		if r.Header.Get("APCA-API-KEY-ID") != "k" {
			t.Errorf("missing auth header")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"trade": map[string]any{"p": 187.42}, "symbol": "AAPL"})
	}))
	defer server.Close()

	b := testBroker(server)
	p, err := b.Price("aapl")
	if err != nil {
		t.Fatalf("Price: %v", err)
	}
	if !p.Equal(decimal.NewFromFloat(187.42)) {
		t.Fatalf("price = %s", p)
	}
}

func TestPriceCrypto(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.EscapedPath() != "/v1beta3/crypto/us/BTC%2FUSD/trades/latest" {
			t.Errorf("unexpected path %s", r.URL.EscapedPath())
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"trade": map[string]any{"p": 50123.5}, "symbol": "BTC/USD"})
	}))
	defer server.Close()

	b := testBroker(server)
	p, err := b.Price("btc/usd")
	if err != nil {
		t.Fatalf("Price: %v", err)
	}
	if !p.Equal(decimal.NewFromFloat(50123.5)) {
		t.Fatalf("price = %s", p)
	}
}


func TestPlaceMarketOrderSendsClientOrderID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["client_order_id"] != "bot1-42" {
			t.Errorf("client_order_id = %q, want bot1-42", body["client_order_id"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "ord-c", "status": "filled", "filled_qty": "1", "filled_avg_price": "100",
		})
	}))
	defer server.Close()

	b := testBroker(server)
	if _, _, _, err := b.PlaceMarketOrder("AAPL", "BUY", decimal.NewFromInt(1), "bot1-42"); err != nil {
		t.Fatalf("PlaceMarketOrder: %v", err)
	}
}

// TestPlaceMarketOrderReconcilesAfterServerError: a 500 on POST must trigger a
// lookup by client_order_id; a filled order is returned instead of an error.
func TestPlaceMarketOrderReconcilesAfterServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/v2/orders" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"message":"server error"}`))
			return
		}
		if r.Method == http.MethodGet && r.URL.Path == "/v2/orders:by_client_order_id" {
			if r.URL.Query().Get("client_order_id") != "bot1-77" {
				t.Errorf("client_order_id = %q", r.URL.Query().Get("client_order_id"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"id": "ord-r", "status": "filled", "filled_qty": "3", "filled_avg_price": "99.5",
			})
			return
		}
		t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	b := testBroker(server)
	avg, qty, _, err := b.PlaceMarketOrder("AAPL", "BUY", decimal.NewFromInt(3), "bot1-77")
	if err != nil {
		t.Fatalf("expected reconciliation, got err: %v", err)
	}
	if !avg.Equal(decimal.NewFromFloat(99.5)) || !qty.Equal(decimal.NewFromInt(3)) {
		t.Fatalf("avg=%s qty=%s", avg, qty)
	}
}

func TestFiltersStocks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/assets/AAPL" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"symbol": "AAPL", "min_order_size": "1", "min_trade_increment": "0.0001", "price_increment": "0.01",
		})
	}))
	defer server.Close()

	b := testBroker(server)
	f, err := b.Filters("aapl")
	if err != nil {
		t.Fatalf("Filters: %v", err)
	}
	if !f.MinQty.Equal(decimal.NewFromInt(1)) || !f.StepQty.Equal(decimal.NewFromFloat(0.0001)) || !f.TickSize.Equal(decimal.NewFromFloat(0.01)) || !f.MinNotional.Equal(decimal.NewFromInt(1)) {
		t.Fatalf("filters = %+v", f)
	}
}

func TestFiltersCryptoMinNotional(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("crypto filters should not call the assets API, got %s", r.URL.Path)
	}))
	defer server.Close()

	b := testBroker(server)
	f, err := b.Filters("BTC/USD")
	if err != nil {
		t.Fatalf("Filters: %v", err)
	}
	if !f.MinNotional.Equal(decimal.NewFromInt(1)) {
		t.Fatalf("crypto min notional = %s, want 1", f.MinNotional)
	}
}
