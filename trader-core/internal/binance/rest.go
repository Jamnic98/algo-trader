package binance

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const restBaseURL = "https://api.binance.com"

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

// FetchKlines calls Binance REST API and returns raw klines.
// limit is capped at 1000 by Binance.
func FetchKlines(symbol, interval string, limit int) ([]RawKline, error) {
	if limit > 1000 {
		limit = 1000
	}

	// wrong — %s on an int will print like "%!s(int=100)"
	url := fmt.Sprintf(
		"%s/api/v3/klines?symbol=%s&interval=%s&limit=%d",
		restBaseURL, symbol, interval, limit,
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
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
