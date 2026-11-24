package alpaca

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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

// PlaceMarketOrder places a market order. Alpaca is commission-free, so fee is 0.
func (b *Broker) PlaceMarketOrder(symbol string, side string, qty decimal.Decimal) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
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
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("alpaca order bad status: %s, body: %s", resp.Status, respBody)
	}

	var order struct {
		ID             string  `json:"id"`
		Status         string  `json:"status"`
		FilledQty      string  `json:"filled_qty"`
		FilledAvgPrice *string `json:"filled_avg_price"`
		RejectReason   string  `json:"reject_reason"`
	}
	if err := json.Unmarshal(respBody, &order); err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("decode order: %w", err)
	}

	if order.Status == "rejected" || order.RejectReason != "" {
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("alpaca order rejected: %s", order.RejectReason)
	}

	filledQty, err := decimal.NewFromString(order.FilledQty)
	if err != nil || filledQty.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("order not filled (status %s)", order.Status)
	}
	if order.FilledAvgPrice == nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("no filled_avg_price on order %s", order.ID)
	}
	avgPrice, err := decimal.NewFromString(*order.FilledAvgPrice)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, err
	}

	// Alpaca charges no commission on stocks/crypto spot
	return avgPrice, filledQty, decimal.Zero, nil
}
