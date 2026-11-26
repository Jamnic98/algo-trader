package alpaca

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"trader-core/internal/exchange"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

// timeframe maps our interval strings to Alpaca bar timeframes.
func timeframe(interval string) (string, error) {
	switch interval {
	case "1m":
		return "1Min", nil
	case "5m":
		return "5Min", nil
	case "15m":
		return "15Min", nil
	case "30m":
		return "30Min", nil
	case "1h":
		return "1Hour", nil
	case "4h":
		return "4Hour", nil
	case "1d":
		return "1Day", nil
	case "1w":
		return "1Week", nil
	default:
		return "", fmt.Errorf("unsupported interval %q for alpaca", interval)
	}
}

type bar struct {
	T string      `json:"t"` // RFC3339
	O json.Number `json:"o"`
	H json.Number `json:"h"`
	L json.Number `json:"l"`
	C json.Number `json:"c"`
	V json.Number `json:"v"`
}

func (b bar) kline() (exchange.Kline, error) {
	t, err := time.Parse(time.RFC3339, b.T)
	if err != nil {
		return exchange.Kline{}, fmt.Errorf("bad bar time %q: %w", b.T, err)
	}
	parse := func(n json.Number) (decimal.Decimal, error) {
		return decimal.NewFromString(n.String())
	}
	o, e1 := parse(b.O)
	h, e2 := parse(b.H)
	l, e3 := parse(b.L)
	c, e4 := parse(b.C)
	v, e5 := parse(b.V)
	for _, e := range []error{e1, e2, e3, e4, e5} {
		if e != nil {
			return exchange.Kline{}, e
		}
	}
	return exchange.Kline{
		OpenTime:  t.UnixMilli(),
		CloseTime: t.Add(time.Minute).UnixMilli() - 1, // approximate; bars are timestamped at open
		Open:      o, High: h, Low: l, Close: c, Volume: v,
	}, nil
}

func barsPath(symbol string) string {
	if IsCryptoSymbol(symbol) {
		return "/v1beta3/crypto/us/" + url.PathEscape(symbol) + "/bars"
	}
	return "/v2/stocks/" + url.PathEscape(symbol) + "/bars"
}

// Klines fetches historical bars, paginating with page_token.
func (b *Broker) Klines(symbol, interval string, startMs, endMs int64, limit int) ([]exchange.Kline, error) {
	tf, err := timeframe(interval)
	if err != nil {
		return nil, err
	}
	symbol = NormalizeSymbol(symbol)
	if limit <= 0 {
		limit = 10000
	}

	out := make([]exchange.Kline, 0, limit)
	pageToken := ""

	for len(out) < limit {
		params := url.Values{}
		params.Set("timeframe", tf)
		if startMs > 0 {
			params.Set("start", time.UnixMilli(startMs).UTC().Format(time.RFC3339))
		}
		if endMs > 0 {
			params.Set("end", time.UnixMilli(endMs).UTC().Format(time.RFC3339))
		}
		params.Set("limit", strconv.Itoa(min(limit-len(out), 10000)))
		if pageToken != "" {
			params.Set("page_token", pageToken)
		}
		if !IsCryptoSymbol(symbol) {
			params.Set("feed", "iex") // free tier
		}

		req, err := b.newRequest(http.MethodGet, b.cfg.DataURL+barsPath(symbol)+"?"+params.Encode(), nil)
		if err != nil {
			return nil, err
		}
		resp, err := b.client.Do(req)
		if err != nil {
			return nil, err
		}
		var page struct {
			Bars          []bar  `json:"bars"`
			NextPageToken string `json:"next_page_token"`
		}
		err = json.NewDecoder(resp.Body).Decode(&page)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("decode bars: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("alpaca bars bad status: %s", resp.Status)
		}

		for _, rb := range page.Bars {
			k, err := rb.kline()
			if err != nil {
				return nil, err
			}
			out = append(out, k)
		}

		if page.NextPageToken == "" || len(page.Bars) == 0 {
			break
		}
		pageToken = page.NextPageToken
	}

	return out, nil
}

// StreamKlines subscribes to Alpaca bar updates over websocket. A bar is
// emitted once the *next* bar for the symbol arrives (i.e. it is closed).
func (b *Broker) StreamKlines(ctx context.Context, symbol, interval string, onKline func(exchange.Kline)) error {
	return b.StreamKlineEvents(ctx, symbol, interval, onKline, nil)
}

// StreamKlineEvents implements exchange.TickStreamer: every incoming bar update
// fires onTick (in-progress bar); a bar fires onKline once the next bar arrives.
func (b *Broker) StreamKlineEvents(ctx context.Context, symbol, interval string, onKline, onTick func(exchange.Kline)) error {
	symbol = NormalizeSymbol(symbol)
	wsURL := b.cfg.WSURL
	if IsCryptoSymbol(symbol) && wsURL == defaultWSStocks {
		wsURL = defaultWSCrypto
	}

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return fmt.Errorf("alpaca ws dial: %w", err)
	}
	defer conn.Close()

	auth := map[string]string{"action": "auth", "key": b.cfg.APIKey, "secret": b.cfg.Secret}
	if err := conn.WriteJSON(auth); err != nil {
		return err
	}
	sub := map[string]any{"action": "subscribe", "bars": []string{symbol}}
	if err := conn.WriteJSON(sub); err != nil {
		return err
	}

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-done:
		}
	}()
	defer close(done)

	var last *bar
	for {
		var msgs []map[string]any
		if err := conn.ReadJSON(&msgs); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("alpaca ws read: %w", err)
		}
		for _, m := range msgs {
			t, _ := m["T"].(string)
			if t == "error" {
				return fmt.Errorf("alpaca ws error: %v", m["msg"])
			}
			if t != "b" {
				continue // auth success etc.
			}
			raw, _ := json.Marshal(m)
			var rb bar
			if err := json.Unmarshal(raw, &rb); err != nil {
				continue
			}
			if last != nil && last.T != rb.T {
				if k, err := last.kline(); err == nil && onKline != nil {
					onKline(k)
				}
			}
			if k, err := rb.kline(); err == nil && onTick != nil {
				onTick(k)
			}
			last = &rb
		}
	}
}
