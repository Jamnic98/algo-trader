// Package exchange abstracts trading venues (Binance, Alpaca, ...) behind a
// single Broker interface plus a name-keyed registry. Each venue lives in its
// own package (e.g. internal/binance) and registers itself in init().
package exchange

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"trader-core/internal/engine"

	"github.com/shopspring/decimal"
)

// Kline is a venue-neutral OHLCV candle.
type Kline struct {
	OpenTime  int64
	CloseTime int64
	Open      decimal.Decimal
	High      decimal.Decimal
	Low       decimal.Decimal
	Close     decimal.Decimal
	Volume    decimal.Decimal
}

// Broker is a trading venue: historical + live market data, balances, orders.
// side is "BUY" or "SELL".
type Broker interface {
	Name() string

	// Klines returns klines with openTime in [startMs, endMs] (0 = unbounded
	// on that side), at most `limit`, oldest first.
	Klines(symbol, interval string, startMs, endMs int64, limit int) ([]Kline, error)

	// StreamKlines streams closed klines until ctx is cancelled.
	StreamKlines(ctx context.Context, symbol, interval string, onKline func(Kline)) error

	// Balances returns free balances per asset.
	Balances() (map[string]decimal.Decimal, error)

	// Price returns the latest trade price for the symbol.
	Price(symbol string) (decimal.Decimal, error)

	// Filters returns the venue's trading rules for the symbol (lot size,
	// tick size, min notional) so orders can be validated client-side.
	Filters(symbol string) (engine.SymbolFilters, error)

	// PlaceMarketOrder places a market order; returns avg price, executed qty, fee.
	// clientOrderID is the idempotency key (reconciled on transport errors).
	PlaceMarketOrder(symbol string, side string, qty decimal.Decimal, clientOrderID string) (avgPrice, executedQty, fee decimal.Decimal, err error)
}

// TickStreamer is an optional extension for brokers whose WS delivers every
// bar update (not just closed bars): onTick fires for in-progress bars,
// onKline for closed ones. Brokers without it still stream closed klines only.
type TickStreamer interface {
	StreamKlineEvents(ctx context.Context, symbol, interval string, onKline, onTick func(Kline)) error
}

// Config holds per-venue credentials/URLs, typically from env.
type Config struct {
	APIKey  string
	Secret  string
	BaseURL string // trading REST (orders/account)
	DataURL string // market-data REST (bars/klines)
	WSURL   string // websocket market data
	Paper   bool
}

// ConfigFromEnv builds a Config from <NAME>_API_KEY / <NAME>_SECRET_KEY /
// <NAME>_BASE_URL / <NAME>_WS_URL / <NAME>_PAPER (NAME uppercased, e.g. ALPACA_).
// The secret also falls back to <NAME>_SECRET (Alpaca documents ALPACA_SECRET;
// Binance uses BINANCE_SECRET_KEY).
func ConfigFromEnv(name string) Config {
	p := strings.ToUpper(name)
	secret := os.Getenv(p + "_SECRET_KEY")
	if secret == "" {
		secret = os.Getenv(p + "_SECRET")
	}
	return Config{
		APIKey:  os.Getenv(p + "_API_KEY"),
		Secret:  secret,
		BaseURL: os.Getenv(p + "_BASE_URL"),
		DataURL: os.Getenv(p + "_DATA_URL"),
		WSURL:   os.Getenv(p + "_WS_URL"),
		Paper:   os.Getenv(p+"_PAPER") == "true",
	}
}

// Factory builds a Broker from a Config (applies venue defaults when empty).
type Factory func(cfg Config) (Broker, error)

var (
	mu       sync.RWMutex
	registry = map[string]Factory{}
)

// Register adds a venue to the registry. Called from each venue package's init().
func Register(name string, f Factory) {
	mu.Lock()
	defer mu.Unlock()
	registry[strings.ToLower(name)] = f
}

// New builds the named broker, pulling credentials from env when cfg fields are empty.
func New(name string, cfg Config) (Broker, error) {
	mu.RLock()
	f, ok := registry[strings.ToLower(name)]
	mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown exchange %q (registered: %v)", name, Names())
	}
	if cfg.APIKey == "" || cfg.Secret == "" {
		env := ConfigFromEnv(name)
		if cfg.APIKey == "" {
			cfg.APIKey = env.APIKey
		}
		if cfg.Secret == "" {
			cfg.Secret = env.Secret
		}
		if cfg.BaseURL == "" {
			cfg.BaseURL = env.BaseURL
		}
		if cfg.WSURL == "" {
			cfg.WSURL = env.WSURL
		}
	}
	return f(cfg)
}

// Names returns the registered venue names.
func Names() []string {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]string, 0, len(registry))
	for n := range registry {
		out = append(out, n)
	}
	return out
}

// Known reports whether a venue name is registered.
func Known(name string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := registry[strings.ToLower(name)]
	return ok
}

// SymbolMapper builds the venue-native symbol from a base/quote pair and asset
// type (e.g. binance "BTC"+"USDT" -> "BTCUSDT", alpaca crypto -> "BTC/USD",
// alpaca stocks -> "AAPL").
type SymbolMapper func(base, quote, assetType string) string

var symbolMappers = map[string]SymbolMapper{}

// RegisterSymbolMapper registers how a venue maps base/quote to its native symbol.
func RegisterSymbolMapper(name string, m SymbolMapper) {
	mu.Lock()
	defer mu.Unlock()
	symbolMappers[strings.ToLower(name)] = m
}

// SymbolFor maps base/quote to the venue-native symbol. Falls back to
// uppercased base+quote for venues without a registered mapper.
func SymbolFor(exchangeName, base, quote, assetType string) string {
	mu.RLock()
	m, ok := symbolMappers[strings.ToLower(exchangeName)]
	mu.RUnlock()
	if !ok {
		return strings.ToUpper(base + quote)
	}
	return m(base, quote, assetType)
}

// quoteAssets maps venue -> its quote/cash asset ("USDT", "USD", ...).
var quoteAssets = map[string]string{}

// RegisterQuoteAsset registers the venue's quote (cash) asset.
func RegisterQuoteAsset(name, asset string) {
	mu.Lock()
	defer mu.Unlock()
	quoteAssets[strings.ToLower(name)] = asset
}

// QuoteAssetFor returns the venue's quote asset, defaulting to USDT.
func QuoteAssetFor(name string) string {
	mu.RLock()
	defer mu.RUnlock()
	if a, ok := quoteAssets[strings.ToLower(name)]; ok {
		return a
	}
	return "USDT"
}
