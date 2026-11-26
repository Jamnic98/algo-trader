package binance

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"trader-core/internal/exchange"

	"github.com/shopspring/decimal"
)

func TestBrokerRegisteredAsBinance(t *testing.T) {
	b, err := exchange.New("binance", exchange.Config{})
	if err != nil {
		t.Fatalf("New(binance): %v", err)
	}
	if b.Name() != "binance" {
		t.Fatalf("name = %q", b.Name())
	}
}

func TestBrokerKlinesViaAdapter(t *testing.T) {
	server := klineServer(t, [][]any{
		klineJSON(1_000_000, "100"),
		klineJSON(1_060_000, "101"),
	})
	defer server.Close()
	withServer(t, server)

	b := NewBroker(exchange.Config{})

	// limit-only path
	kl, err := b.Klines("BTCUSDT", "1m", 0, 0, 500)
	if err != nil || len(kl) != 2 {
		t.Fatalf("Klines limit: %v %v", kl, err)
	}
	if !kl[0].Close.Equal(decimal.NewFromInt(100)) || kl[0].OpenTime != 1_000_000 {
		t.Fatalf("kline = %+v", kl[0])
	}

	// range path
	kl, err = b.Klines("BTCUSDT", "1m", 1_000_000, 1_060_000, 0)
	if err != nil || len(kl) != 2 {
		t.Fatalf("Klines range: %v %v", kl, err)
	}
}

func TestBrokerBalancesAndOrdersRequireCreds(t *testing.T) {
	b := NewBroker(exchange.Config{}) // no keys
	if _, err := b.Balances(); err == nil {
		t.Fatal("expected error without credentials")
	}
	if _, _, _, err := b.PlaceMarketOrder("BTCUSDT", "BUY", decimal.NewFromInt(1)); err == nil {
		t.Fatal("expected error without credentials")
	}
}

func TestBrokerBalancesWithCreds(t *testing.T) {
	t.Setenv("BINANCE_API_KEY", "k")
	t.Setenv("BINANCE_SECRET_KEY", "s")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v3/account" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"balances": []map[string]string{{"asset": "USDT", "free": "42.5"}},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	withServer(t, server)

	b := NewBroker(exchange.Config{APIKey: "k", Secret: "s"})
	balances, err := b.Balances()
	if err != nil {
		t.Fatalf("Balances: %v", err)
	}
	if !balances["USDT"].Equal(decimal.NewFromFloat(42.5)) {
		t.Fatalf("balances = %v", balances)
	}
}

func TestBrokerStreamKlinesContextCancel(t *testing.T) {
	b := NewBroker(exchange.Config{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately: StreamKlines should return promptly
	err := b.StreamKlines(ctx, "BTCUSDT", "1m", func(exchange.Kline) {})
	if err == nil {
		t.Fatal("expected ctx error")
	}
}


func TestBrokerPrice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/ticker/price" || r.URL.Query().Get("symbol") != "BTCUSDT" {
			t.Errorf("unexpected %s?%s", r.URL.Path, r.URL.RawQuery)
		}
		_, _ = w.Write([]byte(`{"symbol":"BTCUSDT","price":"50123.50"}`))
	}))
	defer server.Close()
	withServer(t, server)

	b := NewBroker(exchange.Config{})
	p, err := b.Price("btc/usdt")
	if err != nil {
		t.Fatalf("Price: %v", err)
	}
	if !p.Equal(decimal.NewFromFloat(50123.50)) {
		t.Fatalf("price = %s", p)
	}
}
