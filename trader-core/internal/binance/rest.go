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
	const maxPerRequest = 1000

	if limit <= maxPerRequest {
		return fetchKlinesBatch(symbol, interval, limit, 0)
	}

	var all []RawKline
	endTime := int64(0)

	for limit > 0 {
		batch := min(limit, maxPerRequest)
		klines, err := fetchKlinesBatch(symbol, interval, batch, endTime)
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

func fetchKlinesBatch(symbol, interval string, limit int, endTime int64) ([]RawKline, error) {
	url := fmt.Sprintf(
		"%s/api/v3/klines?symbol=%s&interval=%s&limit=%d",
		restBaseURL, symbol, interval, limit,
	)
	if endTime > 0 {
		url += fmt.Sprintf("&endTime=%d", endTime)
	}

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
