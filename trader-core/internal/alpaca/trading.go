package alpaca

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// newRequest builds an authenticated Alpaca request (APCA key headers).
func (b *Broker) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("APCA-API-KEY-ID", b.cfg.APIKey)
	req.Header.Set("APCA-API-SECRET-KEY", b.cfg.Secret)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (b *Broker) requireCreds() error {
	if b.cfg.APIKey == "" || b.cfg.Secret == "" {
		return fmt.Errorf("ALPACA_API_KEY / ALPACA_SECRET not set")
	}
	return nil
}

// Balances returns the account's cash (as "USD") plus open positions.
func (b *Broker) Balances() (map[string]decimal.Decimal, error) {
	if err := b.requireCreds(); err != nil {
		return nil, err
	}

	req, err := b.newRequest(http.MethodGet, b.cfg.BaseURL+"/v2/account", nil)
	if err != nil {
		return nil, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("alpaca account bad status: %s, body: %s", resp.Status, body)
	}
	var account struct {
		Cash string `json:"cash"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, err
	}

	out := map[string]decimal.Decimal{}
	if cash, err := decimal.NewFromString(account.Cash); err == nil && cash.GreaterThan(decimal.Zero) {
		out["USD"] = cash
	}

	preq, err := b.newRequest(http.MethodGet, b.cfg.BaseURL+"/v2/positions", nil)
	if err != nil {
		return nil, err
	}
	presp, err := b.client.Do(preq)
	if err != nil {
		return nil, err
	}
	defer presp.Body.Close()
	if presp.StatusCode == http.StatusOK {
		var positions []struct {
			Symbol string `json:"symbol"`
			Qty    string `json:"qty"`
		}
		if err := json.NewDecoder(presp.Body).Decode(&positions); err == nil {
			for _, p := range positions {
				if q, err := decimal.NewFromString(p.Qty); err == nil && q.GreaterThan(decimal.Zero) {
					out[p.Symbol] = q
				}
			}
		}
	}

	return out, nil
}

// Price returns the latest trade price: stocks via /v2/stocks/{sym}/trades/latest
// (iex feed), crypto via /v1beta3/crypto/us/{sym}/trades/latest.
func (b *Broker) Price(symbol string) (decimal.Decimal, error) {
	if err := b.requireCreds(); err != nil {
		return decimal.Zero, err
	}
	symbol = NormalizeSymbol(symbol)

	path := "/v2/stocks/" + url.PathEscape(symbol) + "/trades/latest"
	params := url.Values{"feed": {"iex"}}
	if IsCryptoSymbol(symbol) {
		path = "/v1beta3/crypto/us/" + url.PathEscape(symbol) + "/trades/latest"
		params = nil
	}

	req, err := b.newRequest(http.MethodGet, b.cfg.DataURL+path+"?"+params.Encode(), nil)
	if err != nil {
		return decimal.Zero, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return decimal.Zero, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return decimal.Zero, fmt.Errorf("alpaca latest trade bad status: %s, body: %s", resp.Status, body)
	}
	var latest struct {
		Trade struct {
			P json.Number `json:"p"`
		} `json:"trade"`
	}
	if err := json.Unmarshal(body, &latest); err != nil {
		return decimal.Zero, fmt.Errorf("decode latest trade: %w", err)
	}
	return decimal.NewFromString(latest.Trade.P.String())
}

// order is the subset of /v2/orders fields we need.
type order struct {
	ID             string  `json:"id"`
	Status         string  `json:"status"`
	FilledQty      string  `json:"filled_qty"`
	FilledAvgPrice *string `json:"filled_avg_price"`
	RejectReason   string  `json:"reject_reason"`
}

// terminalOrder reports whether an order status is final.
func terminalOrder(status string) bool {
	switch status {
	case "filled", "canceled", "expired", "rejected", "done_for_day":
		return true
	}
	return false
}

// getOrder fetches the current state of an order.
func (b *Broker) getOrder(id string) (order, error) {
	var o order
	req, err := b.newRequest(http.MethodGet, b.cfg.BaseURL+"/v2/orders/"+url.PathEscape(id), nil)
	if err != nil {
		return o, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return o, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return o, fmt.Errorf("alpaca get order bad status: %s, body: %s", resp.Status, body)
	}
	if err := json.Unmarshal(body, &o); err != nil {
		return o, fmt.Errorf("decode order: %w", err)
	}
	return o, nil
}

// PlaceMarketOrder places a market order and waits until it reaches a terminal
// state (Alpaca accepts market orders asynchronously, so the POST alone often
// returns status "accepted" with filled_qty 0). Alpaca is commission-free, so
// fee is 0. Polls at pollInterval up to pollTimeout.
func (b *Broker) PlaceMarketOrder(symbol string, side string, qty decimal.Decimal) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	return b.placeMarketOrder(symbol, side, qty, 500*time.Millisecond, 15*time.Second)
}

func (b *Broker) placeMarketOrder(symbol string, side string, qty decimal.Decimal, pollInterval, pollTimeout time.Duration) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	if err := b.requireCreds(); err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, err
	}
	symbol = NormalizeSymbol(symbol)

	payload := map[string]string{
		"symbol":        symbol,
		"qty":           qty.String(),
		"side":          strings.ToLower(side),
		"type":          "market",
		"time_in_force": "day",
	}
	body, _ := json.Marshal(payload)

	req, err := b.newRequest(http.MethodPost, b.cfg.BaseURL+"/v2/orders", bytes.NewReader(body))
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, err
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, err
	}
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("alpaca order bad status: %s, body: %s", resp.Status, respBody)
	}

	var o order
	if err := json.Unmarshal(respBody, &o); err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("decode order: %w", err)
	}

	// poll until the order reaches a terminal state
	deadline := time.Now().Add(pollTimeout)
	for !terminalOrder(o.Status) {
		if time.Now().After(deadline) {
			return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("order %s not terminal after %s (status %s)", o.ID, pollTimeout, o.Status)
		}
		time.Sleep(pollInterval)
		next, err := b.getOrder(o.ID)
		if err != nil {
			return decimal.Zero, decimal.Zero, decimal.Zero, err
		}
		o = next
	}

	if o.Status == "rejected" || o.RejectReason != "" {
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("alpaca order rejected: %s", o.RejectReason)
	}

	filledQty, err := decimal.NewFromString(o.FilledQty)
	if err != nil || filledQty.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("order %s not filled (status %s)", o.ID, o.Status)
	}
	if o.FilledAvgPrice == nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("no filled_avg_price on order %s", o.ID)
	}
	avgPrice, err := decimal.NewFromString(*o.FilledAvgPrice)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, err
	}

	// Alpaca charges no commission on stocks/crypto spot
	return avgPrice, filledQty, decimal.Zero, nil
}
