package strategies

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"trader-core/internal/db/models"

	"github.com/shopspring/decimal"
)

func httpCandle(close int64) CandleContext {
	p := decimal.NewFromInt(close)
	return CandleContext{
		Candle:  models.Candle{Close: p, Open: p, High: p, Low: p},
		FeeRate: decimal.NewFromFloat(0.001),
	}
}

func TestHTTPStrategySignalsAndWindow(t *testing.T) {
	var lastBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&lastBody); err != nil {
			t.Errorf("decode: %v", err)
		}
		closes := lastBody["closes"].([]any)
		if len(closes) < 2 {
			_ = json.NewEncoder(w).Encode(map[string]any{"signal": "HOLD"})
			return
		}
		last, _ := strconv.ParseFloat(closes[len(closes)-1].(string), 64)
		prev, _ := strconv.ParseFloat(closes[len(closes)-2].(string), 64)
		sig := "HOLD"
		if last > prev {
			sig = "BUY"
		} else if last < prev {
			sig = "SELL"
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"signal": sig, "quantity": "0.5"})
	}))
	defer server.Close()

	s, err := NewHTTPStrategy(map[string]any{"url": server.URL + "/predict", "window": float64(3), "model": "toy"}, nil)
	if err != nil {
		t.Fatalf("NewHTTPStrategy: %v", err)
	}

	// up move -> BUY with quantity 0.5
	d := s.OnCandle(httpCandle(100))
	if d.Signal != Hold {
		t.Fatalf("first candle: %v", d.Signal)
	}
	d = s.OnCandle(httpCandle(101))
	if d.Signal != Buy || d.Quantity == nil || !d.Quantity.Equal(decimal.NewFromFloat(0.5)) {
		t.Fatalf("up candle: %+v, want BUY qty 0.5", d)
	}

	// window is capped at 3 closes, extra params forwarded
	d = s.OnCandle(httpCandle(102))
	closes := lastBody["closes"].([]any)
	if len(closes) != 3 {
		t.Fatalf("window = %d, want 3", len(closes))
	}
	params := lastBody["params"].(map[string]any)
	if params["model"] != "toy" {
		t.Fatalf("extra params not forwarded: %v", params)
	}
	if _, leaked := params["url"]; leaked {
		t.Fatal("url should not be forwarded in params")
	}
	_ = d

	// down move -> SELL
	if d := s.OnCandle(httpCandle(90)); d.Signal != Sell {
		t.Fatalf("down candle: %v, want SELL", d.Signal)
	}
}

func TestHTTPStrategyFailuresHold(t *testing.T) {
	// unreachable server
	s, err := NewHTTPStrategy(map[string]any{"url": "http://127.0.0.1:1/predict"}, nil)
	if err != nil {
		t.Fatalf("NewHTTPStrategy: %v", err)
	}
	if d := s.OnCandle(httpCandle(100)); d.Signal != Hold || d.Reason == "" {
		t.Fatalf("unreachable: %+v, want HOLD with reason", d)
	}

	// invalid signal from server
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"signal": "YOLO"})
	}))
	defer bad.Close()
	s2, _ := NewHTTPStrategy(map[string]any{"url": bad.URL}, nil)
	if d := s2.OnCandle(httpCandle(100)); d.Signal != Hold {
		t.Fatalf("invalid signal: %v, want HOLD", d.Signal)
	}

	// missing url param
	if _, err := NewHTTPStrategy(map[string]any{}, nil); err == nil {
		t.Fatal("expected error when url param missing")
	}
}

func TestHTTPTemplateRegistered(t *testing.T) {
	s, err := Build("ml_example", []byte(`{"url":"http://127.0.0.1:1/predict"}`))
	if err != nil {
		t.Fatalf("Build(ml_example): %v", err)
	}
	if _, ok := s.(*HTTPStrategy); !ok {
		t.Fatalf("got %T, want *HTTPStrategy", s)
	}
	// defaults from schema applied (window 20)
	if s.(*HTTPStrategy).window != 20 {
		t.Fatalf("window = %d, want 20 (schema default)", s.(*HTTPStrategy).window)
	}
}
