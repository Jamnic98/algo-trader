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

	// PlaceMarketOrder places a market order; returns avg price, executed qty, fee.
	PlaceMarketOrder(symbol string, side string, qty decimal.Decimal) (avgPrice, executedQty, fee decimal.Decimal, err error)
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
func ConfigFromEnv(name string) Config {
	p := strings.ToUpper(name)
	return Config{
		APIKey:  os.Getenv(p + "_API_KEY"),
		Secret:  os.Getenv(p + "_SECRET"),
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
