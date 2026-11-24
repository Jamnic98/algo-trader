package binance

import (
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
	"trader-core/internal/engine"
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

// place a market trade (signed endpoint, credentials from env)
func PlaceTrade(symbol string, side engine.Side, quantity float64) (string, error) {
	params := url.Values{}
	params.Set("symbol", NormalizeSymbol(symbol))
	params.Set("side", string(side))
	params.Set("type", "MARKET")
	if quantity != 0 {
		params.Set("quantity", strconv.FormatFloat(quantity, 'f', -1, 64))
	}

	req, err := signedRequest(http.MethodPost, "/api/v3/order", params)
	if err != nil {
		return "", fmt.Errorf("place trade request failed: %w", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("place trade request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("place trade bad status: %s, body: %s", resp.Status, body)
	}

	fmt.Println("status", resp.StatusCode)

	var result struct {
		Price string `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("place trade response decode failed: %w", err)
	}

	return result.Price, nil
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
