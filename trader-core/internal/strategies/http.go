package strategies

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
)

// HTTPStrategy implements the "http" template: strategy logic lives in an
// external service (e.g. an ML inference endpoint). Each candle, the engine
// POSTs the recent close window plus the configured params and expects
// {"signal": "BUY"|"SELL"|"HOLD", "quantity": "..." (optional)} back.
// Mirrored by research/app/strategies/http.py.
type HTTPStrategy struct {
	url    string
	window int
	extra  map[string]any
	closes []string
	client *http.Client
}

func NewHTTPStrategy(params map[string]any, client *http.Client) (*HTTPStrategy, error) {
	url, _ := params["url"].(string)
	if url == "" {
		return nil, fmt.Errorf("http strategy requires a \"url\" param")
	}
	window := 100
	switch w := params["window"].(type) {
	case float64:
		if w > 0 {
			window = int(w)
		}
	case int:
		if w > 0 {
			window = w
		}
	}
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	extra := map[string]any{}
	for k, v := range params {
		if k != "url" && k != "window" {
			extra[k] = v
		}
	}
	return &HTTPStrategy{url: url, window: window, extra: extra, client: client}, nil
}

func (s *HTTPStrategy) OnCandle(ctx CandleContext) Decision {
	s.closes = append(s.closes, ctx.Candle.Close.String())
	if len(s.closes) > s.window {
		s.closes = s.closes[len(s.closes)-s.window:]
	}

	body, err := json.Marshal(map[string]any{
		"closes": s.closes,
		"params": s.extra,
	})
	if err != nil {
		return Decision{Signal: Hold, Reason: err.Error()}
	}

	req, err := http.NewRequest(http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		return Decision{Signal: Hold, Reason: err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return Decision{Signal: Hold, Reason: fmt.Sprintf("http strategy unreachable: %v", err)}
	}
	defer resp.Body.Close()

	var out struct {
		Signal   string `json:"signal"`
		Quantity string `json:"quantity"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Decision{Signal: Hold, Reason: fmt.Sprintf("bad response: %v", err)}
	}

	switch Signal(out.Signal) {
	case Buy, Sell, Hold:
	default:
		return Decision{Signal: Hold, Reason: fmt.Sprintf("invalid signal %q", out.Signal)}
	}

	var qty *decimal.Decimal
	if out.Quantity != "" {
		q, err := decimal.NewFromString(out.Quantity)
		if err != nil {
			return Decision{Signal: Hold, Reason: fmt.Sprintf("bad quantity %q", out.Quantity)}
		}
		qty = &q
	}

	return Decision{Signal: Signal(out.Signal), Quantity: qty}
}
