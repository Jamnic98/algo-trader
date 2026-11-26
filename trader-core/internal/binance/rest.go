package binance

import (
	"trader-core/internal/engine"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// var (not const) so tests can point it at a local httptest server
var restBaseURL = "https://api.binance.com"

type RawKline struct {
	OpenTime  int64
	Open      string
	High      string
	Low       string
	Close     string
	Volume    string
	CloseTime int64
}

func (r *RawKline) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw) < 7 {
		return fmt.Errorf("kline array too short")
	}

	if err := json.Unmarshal(raw[0], &r.OpenTime); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[1], &r.Open); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[2], &r.High); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[3], &r.Low); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[4], &r.Close); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[5], &r.Volume); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[6], &r.CloseTime); err != nil {
		return err
	}
	return nil
}

// FetchKlines calls Binance REST API and returns the most recent `limit` klines.
// limit is capped at 1000 per request by Binance; larger limits are paginated.
func FetchKlines(symbol, interval string, limit int) ([]RawKline, error) {
	const maxPerRequest = 1000

	if limit <= maxPerRequest {
		return fetchKlinesBatch(symbol, interval, limit, 0, 0)
	}

	var all []RawKline
	endTime := int64(0)

	for limit > 0 {
		batch := min(limit, maxPerRequest)
		klines, err := fetchKlinesBatch(symbol, interval, batch, 0, endTime)
		if err != nil {
			return nil, err
		}
		if len(klines) == 0 {
			break
		}

		all = append(klines, all...) // prepend — oldest first
		endTime = klines[0].OpenTime - 1
		limit -= len(klines)
	}

	return all, nil
}

// FetchKlinesRange fetches klines with openTime in [startTime, endTime] (ms),
// paginating forward in batches of 1000. Used by backtests over a date range.
func FetchKlinesRange(symbol, interval string, startTime, endTime int64) ([]RawKline, error) {
	const maxPerRequest = 1000

	var all []RawKline
	cursor := startTime

	for {
		batch, err := fetchKlinesBatch(symbol, interval, maxPerRequest, cursor, endTime)
		if err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			break
		}

		all = append(all, batch...)

		last := batch[len(batch)-1].OpenTime
		if last >= endTime || len(batch) < maxPerRequest {
			break
		}
		cursor = last + 1
	}

	return all, nil
}

func fetchKlinesBatch(symbol, interval string, limit int, startTime, endTime int64) ([]RawKline, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("interval", interval)
	params.Set("limit", strconv.Itoa(limit))
	if startTime > 0 {
		params.Set("startTime", strconv.FormatInt(startTime, 10))
	}
	if endTime > 0 {
		params.Set("endTime", strconv.FormatInt(endTime, 10))
	}
	reqURL := fmt.Sprintf("%s/api/v3/klines?%s", restBaseURL, params.Encode())

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("FetchKlines request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("FetchKlines bad status: %s, body: %s", resp.Status, body)
	}

	var klines []RawKline
	if err := json.NewDecoder(resp.Body).Decode(&klines); err != nil {
		return nil, fmt.Errorf("FetchKlines decode failed: %w", err)
	}

	return klines, nil
}

// signedRequest builds a signed Binance request. Credentials come from the
// BINANCE_API_KEY / BINANCE_SECRET_KEY environment variables — never hardcode them.
func signedRequest(method, path string, params url.Values) (*http.Request, error) {
	apiKey := os.Getenv("BINANCE_API_KEY")
	secretKey := os.Getenv("BINANCE_SECRET_KEY")
	if apiKey == "" || secretKey == "" {
		return nil, fmt.Errorf("BINANCE_API_KEY / BINANCE_SECRET_KEY not set")
	}

	params.Set("timestamp", strconv.FormatInt(time.Now().UnixMilli(), 10))
	params.Set("recvWindow", "5000")

	mac := hmac.New(sha256.New, []byte(secretKey))
	if _, err := mac.Write([]byte(params.Encode())); err != nil {
		return nil, err
	}
	params.Set("signature", hex.EncodeToString(mac.Sum(nil)))

	req, err := http.NewRequest(method, fmt.Sprintf("%s%s?%s", restBaseURL, path, params.Encode()), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-MBX-APIKEY", apiKey)
	return req, nil
}

func FetchPrice(symbol string) (string, error) {
	url := fmt.Sprintf("%s/api/v3/ticker/price?symbol=%s", restBaseURL, symbol)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("FetchPrice request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("FetchPrice bad status: %s, body: %s", resp.Status, body)
	}

	var result struct {
		Price string `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("FetchPrice decode failed: %w", err)
	}

	return result.Price, nil
}

// NormalizeSymbol converts display formats like "xrp/usdt" or "BTC-USDT"
// into the Binance format "XRPUSDT" (uppercase alphanumerics only).
func NormalizeSymbol(symbol string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(symbol) {
		if (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// SpotPlacer implements engine.OrderPlacer against Binance spot REST
// (signed endpoint, credentials from BINANCE_API_KEY / BINANCE_SECRET_KEY).
type SpotPlacer struct{}

// PlaceMarketOrder places a spot MARKET order and returns the executed
// average price, executed quantity and total commission (fee). clientOrderID
// is sent as newClientOrderId; if the request fails at transport level (or
// 5xx), the order is reconciled by clientOrderID so a filled order is never
// reported as failed.
func (SpotPlacer) PlaceMarketOrder(symbol string, side string, qty decimal.Decimal, clientOrderID string) (avgPrice, executedQty, fee decimal.Decimal, err error) {
	params := url.Values{}
	params.Set("symbol", NormalizeSymbol(symbol))
	params.Set("side", string(side))
	params.Set("type", "MARKET")
	params.Set("quantity", qty.String())
	if clientOrderID != "" {
		params.Set("newClientOrderId", clientOrderID)
	}

	req, err := signedRequest(http.MethodPost, "/api/v3/order", params)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("place order request failed: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		// transport error: the order may have reached the venue — reconcile
		if clientOrderID != "" {
			if a, q, f, rerr := queryOrderByClientID(NormalizeSymbol(symbol), clientOrderID); rerr == nil {
				return a, q, f, nil
			}
		}
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("place order request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode >= 500 && clientOrderID != "" {
			if a, q, f, rerr := queryOrderByClientID(NormalizeSymbol(symbol), clientOrderID); rerr == nil {
				return a, q, f, nil
			}
		}
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("place order bad status: %s, body: %s", resp.Status, body)
	}

	var result struct {
		OrderID             int64  `json:"orderId"`
		Symbol              string `json:"symbol"`
		Side                string `json:"side"`
		Status              string `json:"status"`
		ExecutedQty         string `json:"executedQty"`
		CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
		Fills               []struct {
			Price           string `json:"price"`
			Qty             string `json:"qty"`
			Commission      string `json:"commission"`
			CommissionAsset string `json:"commissionAsset"`
		} `json:"fills"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("place order response decode failed: %w", err)
	}

	executedQty, err = decimal.NewFromString(result.ExecutedQty)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("bad executedQty %q: %w", result.ExecutedQty, err)
	}

	cum, _ := decimal.NewFromString(result.CummulativeQuoteQty)
	if executedQty.GreaterThan(decimal.Zero) && cum.GreaterThan(decimal.Zero) {
		avgPrice = cum.Div(executedQty)
	} else {
		// fall back to summing fills
		var sumQty, sumNotional, sumFee decimal.Decimal
		for _, f := range result.Fills {
			p, _ := decimal.NewFromString(f.Price)
			q, _ := decimal.NewFromString(f.Qty)
			c, _ := decimal.NewFromString(f.Commission)
			sumQty = sumQty.Add(q)
			sumNotional = sumNotional.Add(p.Mul(q))
			sumFee = sumFee.Add(c)
		}
		if sumQty.GreaterThan(decimal.Zero) {
			avgPrice = sumNotional.Div(sumQty)
		}
		fee = sumFee
	}
	if fee.IsZero() {
		for _, f := range result.Fills {
			c, _ := decimal.NewFromString(f.Commission)
			fee = fee.Add(c)
		}
	}

	return avgPrice, executedQty, fee, nil
}

// FetchAccount returns free balances per asset from the signed account endpoint.
func FetchAccount() (map[string]decimal.Decimal, error) {
	req, err := signedRequest(http.MethodGet, "/api/v3/account", url.Values{})
	if err != nil {
		return nil, fmt.Errorf("account request failed: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("account request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("account bad status: %s, body: %s", resp.Status, body)
	}

	var result struct {
		Balances []struct {
			Asset string `json:"asset"`
			Free  string `json:"free"`
		} `json:"balances"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("account response decode failed: %w", err)
	}

	out := make(map[string]decimal.Decimal, len(result.Balances))
	for _, b := range result.Balances {
		free, err := decimal.NewFromString(b.Free)
		if err != nil {
			continue
		}
		if free.GreaterThan(decimal.Zero) {
			out[b.Asset] = free
		}
	}
	return out, nil
}

func CheckServerTime() (int, error) {
	url := fmt.Sprintf("%s/api/v3/time", restBaseURL)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return 0, fmt.Errorf("check time request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("check time bad status: %s, body: %s", resp.Status, body)
	}

	var result struct {
		ServerTime int `json:"serverTime"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("time check decode failed: %w", err)
	}

	return result.ServerTime, nil
}

// queryOrderByClientID fetches an order by its client order id and, if filled,
// returns avg price, executed qty and fee (reconciliation after transport errors).
func queryOrderByClientID(symbol, clientOrderID string) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("origClientOrderId", clientOrderID)
	req, err := signedRequest(http.MethodGet, "/api/v3/order", params)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("query order bad status: %s", resp.Status)
	}
	var result struct {
		Status              string `json:"status"`
		ExecutedQty         string `json:"executedQty"`
		CummulativeQuoteQty string `json:"cummulativeQuoteQty"`
		Fills               []struct {
			Price      string `json:"price"`
			Qty        string `json:"qty"`
			Commission string `json:"commission"`
		} `json:"fills"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, err
	}
	if result.Status != "FILLED" {
		return decimal.Zero, decimal.Zero, decimal.Zero, fmt.Errorf("order not filled (status %s)", result.Status)
	}
	executedQty, err := decimal.NewFromString(result.ExecutedQty)
	if err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, err
	}
	cum, _ := decimal.NewFromString(result.CummulativeQuoteQty)
	var avgPrice, fee decimal.Decimal
	if executedQty.GreaterThan(decimal.Zero) && cum.GreaterThan(decimal.Zero) {
		avgPrice = cum.Div(executedQty)
	}
	for _, f := range result.Fills {
		c, _ := decimal.NewFromString(f.Commission)
		fee = fee.Add(c)
	}
	return avgPrice, executedQty, fee, nil
}

// filtersForSymbol fetches /api/v3/exchangeInfo and extracts LOT_SIZE,
// PRICE_FILTER and MIN_NOTIONAL for the symbol.
func filtersForSymbol(symbol string) (engine.SymbolFilters, error) {
	var f engine.SymbolFilters
	symbol = NormalizeSymbol(symbol)
	resp, err := http.Get(fmt.Sprintf("%s/api/v3/exchangeInfo?symbol=%s", restBaseURL, symbol))
	if err != nil {
		return f, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return f, fmt.Errorf("exchangeInfo bad status: %s", resp.Status)
	}
	var info struct {
		Symbols []struct {
			Symbol   string `json:"symbol"`
			Filters  []struct {
				FilterType string `json:"filterType"`
				MinQty     string `json:"minQty"`
				MaxQty     string `json:"maxQty"`
				StepSize   string `json:"stepSize"`
				MinPrice   string `json:"minPrice"`
				TickSize   string `json:"tickSize"`
				MinNotional string `json:"minNotional"`
			} `json:"filters"`
		} `json:"symbols"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return f, err
	}
	for _, sym := range info.Symbols {
		if sym.Symbol != symbol {
			continue
		}
		for _, fl := range sym.Filters {
			switch fl.FilterType {
			case "LOT_SIZE":
				f.MinQty, _ = decimal.NewFromString(fl.MinQty)
				f.MaxQty, _ = decimal.NewFromString(fl.MaxQty)
				f.StepQty, _ = decimal.NewFromString(fl.StepSize)
			case "PRICE_FILTER":
				f.TickSize, _ = decimal.NewFromString(fl.TickSize)
			case "MIN_NOTIONAL":
				f.MinNotional, _ = decimal.NewFromString(fl.MinNotional)
			}
		}
		return f, nil
	}
	return f, fmt.Errorf("symbol %s not found in exchangeInfo", symbol)
}
