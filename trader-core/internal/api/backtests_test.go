package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseTimeMs(t *testing.T) {
	cases := []struct {
		in      string
		want    int64
		wantErr bool
	}{
		{in: "1700000000000", want: 1700000000000},
		{in: "2024-01-01", want: 1704067200000},
		{in: "2024-01-01T00:00:00Z", want: 1704067200000},
		{in: "not-a-date", wantErr: true},
	}

	for _, tc := range cases {
		got, err := parseTimeMs(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseTimeMs(%q): expected error, got %d", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseTimeMs(%q): unexpected error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("parseTimeMs(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestPostBacktest(t *testing.T) {
	var gotPath string
	var gotPayload map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotPayload)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stats":{"total_return_pct":"1.5"}}`))
	}))
	defer server.Close()

	status, body, err := postBacktest(context.Background(), server.URL+"/", "/backtest", map[string]any{
		"strategy_slug":   "simple",
		"symbol":          "BTCUSDT",
		"interval":        "1h",
		"start":           int64(1704067200000),
		"end":             int64(1706745600000),
		"initial_balance": "10000",
		"fee_rate":        "0.001",
	})
	if err != nil {
		t.Fatalf("postBacktest error: %v", err)
	}
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", status, body)
	}
	if gotPath != "/backtest" {
		t.Errorf("path = %q, want /backtest", gotPath)
	}
	if gotPayload["strategy_slug"] != "simple" || gotPayload["symbol"] != "BTCUSDT" {
		t.Errorf("unexpected payload: %v", gotPayload)
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("response not JSON: %v", err)
	}
	stats, ok := result["stats"].(map[string]any)
	if !ok || stats["total_return_pct"] != "1.5" {
		t.Errorf("unexpected result body: %s", body)
	}
}

func TestPostBacktestUpstreamError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"unknown strategy"}`))
	}))
	defer server.Close()

	status, body, err := postBacktest(context.Background(), server.URL, "/backtest", map[string]any{})
	if err != nil {
		t.Fatalf("postBacktest error: %v", err)
	}
	if status != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", status)
	}
	if string(body) != `{"error":"unknown strategy"}` {
		t.Errorf("body = %s", body)
	}
}

func TestPostBacktestUnreachable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := server.URL
	server.Close() // immediately unreachable

	_, _, err := postBacktest(context.Background(), url, "/backtest", map[string]any{})
	if err == nil {
		t.Fatal("expected error for unreachable backtester")
	}
}
