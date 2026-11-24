// Package alpaca implements exchange.Broker for Alpaca (stocks + crypto,
// paper trading by default). Credentials come from ALPACA_API_KEY /
// ALPACA_SECRET (via exchange.ConfigFromEnv); ALPACA_PAPER=true selects the
// paper trading API.
package alpaca

import (
	"net/http"
	"time"

	"trader-core/internal/exchange"
)

const (
	defaultTradingURLPaper = "https://paper-api.alpaca.markets"
	defaultTradingURLLive  = "https://api.alpaca.markets"
	defaultDataURL         = "https://data.alpaca.markets"
	defaultWSStocks        = "wss://stream.data.alpaca.markets/v2/iex"
	defaultWSCrypto        = "wss://stream.data.alpaca.markets/v1beta1/crypto/us"
)

func init() {
	exchange.Register("alpaca", func(cfg exchange.Config) (exchange.Broker, error) {
		return New(cfg), nil
	})
	exchange.RegisterSymbolMapper("alpaca", func(base, quote, assetType string) string {
		return SymbolFor(base, quote, assetType == "crypto")
	})
}

// Broker is an Alpaca venue (paper or live).
type Broker struct {
	cfg    exchange.Config
	client *http.Client
}

func New(cfg exchange.Config) *Broker {
	if cfg.BaseURL == "" {
		if cfg.Paper {
			cfg.BaseURL = defaultTradingURLPaper
		} else {
			cfg.BaseURL = defaultTradingURLLive
		}
	}
	if cfg.DataURL == "" {
		cfg.DataURL = defaultDataURL
	}
	if cfg.WSURL == "" {
		cfg.WSURL = defaultWSStocks
	}
	return &Broker{cfg: cfg, client: &http.Client{Timeout: 15 * time.Second}}
}

func (b *Broker) Name() string { return "alpaca" }

// SymbolFor builds the Alpaca symbol for a base/quote pair: crypto uses
// "BTC/USD", stocks just "AAPL".
func SymbolFor(base, quote string, crypto bool) string {
	if crypto {
		return upper(base) + "/" + upper(quote)
	}
	return upper(base)
}

// NormalizeSymbol uppercases and keeps the crypto slash.
func NormalizeSymbol(symbol string) string {
	return upper(symbol)
}

func upper(s string) string {
	out := []rune(s)
	for i, r := range out {
		if r >= 'a' && r <= 'z' {
			out[i] = r - 32
		}
	}
	return string(out)
}

// IsCryptoSymbol reports whether the symbol is an Alpaca crypto pair ("BTC/USD").
func IsCryptoSymbol(symbol string) bool {
	for _, r := range symbol {
		if r == '/' {
			return true
		}
	}
	return false
}
