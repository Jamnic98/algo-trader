package exchange

import (
	"context"
	"testing"

	"trader-core/internal/engine"

	"github.com/shopspring/decimal"
)

type fakeBroker struct{ name string }

func (f fakeBroker) Name() string { return f.name }
func (f fakeBroker) Klines(symbol, interval string, startMs, endMs int64, limit int) ([]Kline, error) {
	return []Kline{{OpenTime: startMs, Close: decimal.NewFromInt(1)}}, nil
}
func (f fakeBroker) StreamKlines(ctx context.Context, symbol, interval string, onKline func(Kline)) error {
	return nil
}
func (f fakeBroker) Balances() (map[string]decimal.Decimal, error) {
	return map[string]decimal.Decimal{"USD": decimal.NewFromInt(100)}, nil
}
func (f fakeBroker) Price(symbol string) (decimal.Decimal, error) {
	return decimal.NewFromInt(42), nil
}
func (f fakeBroker) PlaceMarketOrder(symbol string, side string, qty decimal.Decimal, clientOrderID string) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	return decimal.NewFromInt(1), qty, decimal.Zero, nil
}
func (f fakeBroker) Filters(symbol string) (engine.SymbolFilters, error) {
	return engine.SymbolFilters{}, nil
}

func TestRegistryRegisterAndNew(t *testing.T) {
	Register("fake", func(cfg Config) (Broker, error) { return fakeBroker{"fake"}, nil })

	b, err := New("FAKE", Config{}) // case-insensitive lookup
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if b.Name() != "fake" {
		t.Fatalf("name = %q", b.Name())
	}

	kl, err := b.Klines("BTCUSDT", "1h", 100, 200, 10)
	if err != nil || len(kl) != 1 || kl[0].OpenTime != 100 {
		t.Fatalf("klines = %v, %v", kl, err)
	}

	if _, err := New("nope", Config{}); err == nil {
		t.Fatal("expected unknown exchange error")
	}

	found := false
	for _, n := range Names() {
		if n == "fake" {
			found = true
		}
	}
	if !found {
		t.Fatalf("Names() = %v", Names())
	}
}

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("ALPACA_API_KEY", "k")
	t.Setenv("ALPACA_SECRET", "s")
	t.Setenv("ALPACA_BASE_URL", "https://paper-api.alpaca.markets")
	t.Setenv("ALPACA_PAPER", "true")

	cfg := ConfigFromEnv("alpaca")
	if cfg.APIKey != "k" || cfg.Secret != "s" || cfg.BaseURL != "https://paper-api.alpaca.markets" || !cfg.Paper {
		t.Fatalf("cfg = %+v", cfg)
	}
}

func TestSymbolMapperRegistry(t *testing.T) {
	RegisterSymbolMapper("mapperx", func(base, quote, assetType string) string {
		return base + "/" + quote + ":" + assetType
	})
	if got := SymbolFor("MAPPERX", "btc", "usd", "crypto"); got != "btc/usd:crypto" {
		t.Fatalf("SymbolFor = %q", got)
	}
	// fallback for a venue without a registered mapper
	if got := SymbolFor("unregistered-venue", "btc", "usdt", "crypto"); got != "BTCUSDT" {
		t.Fatalf("fallback SymbolFor = %q", got)
	}
}

func TestKnown(t *testing.T) {
	Register("knownx", func(cfg Config) (Broker, error) { return fakeBroker{"knownx"}, nil })
	if !Known("KNOWNX") {
		t.Fatal("Known(KNOWNX) = false, want true")
	}
	if Known("definitely-not-registered") {
		t.Fatal("Known = true for unknown venue")
	}
}

func TestQuoteAssetRegistry(t *testing.T) {
	// unregistered venue falls back to USDT
	if got := QuoteAssetFor("definitely-not-registered"); got != "USDT" {
		t.Fatalf("QuoteAssetFor(unknown) = %q, want USDT", got)
	}
	RegisterQuoteAsset("quotex", "USD")
	if got := QuoteAssetFor("QUOTEX"); got != "USD" { // case-insensitive
		t.Fatalf("QuoteAssetFor(QUOTEX) = %q, want USD", got)
	}
}
