package binance

import (
	"context"
	"fmt"
	"strings"

	"trader-core/internal/engine"
	"trader-core/internal/exchange"

	"github.com/shopspring/decimal"
)

func init() {
	exchange.Register("binance", func(cfg exchange.Config) (exchange.Broker, error) {
		return NewBroker(cfg), nil
	})
	exchange.RegisterSymbolMapper("binance", func(base, quote, _ string) string {
		return NormalizeSymbol(base + quote)
	})
}

// Broker adapts the existing Binance REST/WS code to exchange.Broker.
type Broker struct {
	cfg exchange.Config
}

func NewBroker(cfg exchange.Config) *Broker {
	if cfg.BaseURL != "" {
		restBaseURL = cfg.BaseURL
	}
	if cfg.WSURL != "" {
		binanceWSURL = cfg.WSURL
	}
	return &Broker{cfg: cfg}
}

func (b *Broker) Name() string { return "binance" }

func (b *Broker) Klines(symbol, interval string, startMs, endMs int64, limit int) ([]exchange.Kline, error) {
	var raw []RawKline
	var err error
	if startMs > 0 {
		raw, err = FetchKlinesRange(symbol, interval, startMs, endMs)
	} else {
		if limit <= 0 {
			limit = 1000
		}
		raw, err = FetchKlines(symbol, interval, limit)
	}
	if err != nil {
		return nil, err
	}
	out := make([]exchange.Kline, 0, len(raw))
	for _, k := range raw {
		out = append(out, exchange.Kline{
			OpenTime:  k.OpenTime,
			CloseTime: k.CloseTime,
			Open:      decimal.RequireFromString(k.Open),
			High:      decimal.RequireFromString(k.High),
			Low:       decimal.RequireFromString(k.Low),
			Close:     decimal.RequireFromString(k.Close),
			Volume:    decimal.RequireFromString(k.Volume),
		})
	}
	return out, nil
}

func (b *Broker) StreamKlines(ctx context.Context, symbol, interval string, onKline func(exchange.Kline)) error {
	client := NewClient(ctx)
	if err := client.StartWithReconnect(func(msg string) {}); err != nil {
		return err
	}
	client.Subscribe(strings.ToLower(symbol) + "@kline_" + interval)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-client.Messages():
			candle, _, ok := engine.ParseKline(msg)
			if !ok {
				continue
			}
			onKline(exchange.Kline{
				OpenTime:  candle.OpenTime,
				CloseTime: candle.CloseTime,
				Open:      candle.Open,
				High:      candle.High,
				Low:       candle.Low,
				Close:     candle.Close,
				Volume:    candle.Volume,
			})
		}
	}
}

func (b *Broker) Balances() (map[string]decimal.Decimal, error) {
	if b.cfg.APIKey == "" || b.cfg.Secret == "" {
		return nil, fmt.Errorf("BINANCE_API_KEY / BINANCE_SECRET_KEY not set")
	}
	return FetchAccount()
}

func (b *Broker) PlaceMarketOrder(symbol string, side string, qty decimal.Decimal) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	if b.cfg.APIKey == "" || b.cfg.Secret == "" {
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("BINANCE_API_KEY / BINANCE_SECRET_KEY not set")
	}
	return SpotPlacer{}.PlaceMarketOrder(symbol, side, qty)
}
