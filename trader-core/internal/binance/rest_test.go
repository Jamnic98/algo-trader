package binance

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"trader-core/internal/engine"
)

func klineJSON(openTime int64, close string) []any {
	return []any{openTime, close, close, close, close, "1.0", openTime + 59_999, "0", 0, "0", "0", "0"}
}

// klineServer serves klines from a fixed ascending slice, honoring
// startTime/endTime/limit like Binance does.
func klineServer(t *testing.T, klines [][]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/klines" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		limit, _ := strconv.Atoi(q.Get("limit"))
		var start, end int64
		if s := q.Get("startTime"); s != "" {
			start, _ = strconv.ParseInt(s, 10, 64)
		}
		if e := q.Get("endTime"); e != "" {
			end, _ = strconv.ParseInt(e, 10, 64)
		}

		var out [][]any
		for _, k := range klines {
			ot := k[0].(int64)
			if start > 0 && ot < start {
				continue
			}
			if end > 0 && ot > end {
				continue
			}
			out = append(out, k)
		}
		// Binance returns ascending; when no startTime, the latest `limit`
		if start == 0 && len(out) > limit {
			out = out[len(out)-limit:]
		}
		if len(out) > limit {
			out = out[:limit]
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}))
}

func withServer(t *testing.T, server *httptest.Server) {
	t.Helper()
	orig := restBaseURL
	restBaseURL = server.URL
	t.Cleanup(func() { restBaseURL = orig })
}

func TestFetchKlinesSingleBatch(t *testing.T) {
	klines := [][]any{klineJSON(1000, "100"), klineJSON(2000, "101")}
	server := klineServer(t, klines)
	defer server.Close()
	withServer(t, server)

	got, err := FetchKlines("BTCUSDT", "1m", 500)
	if err != nil {
		t.Fatalf("FetchKlines: %v", err)
	}
	if len(got) != 2 || got[0].OpenTime != 1000 || got[1].Close != "101" {
		t.Fatalf("got %+v", got)
	}
}

func TestFetchKlinesPaginatesBackward(t *testing.T) {
	var klines [][]any
	for i := int64(0); i < 2500; i++ {
		klines = append(klines, klineJSON(1_000_000+i*60_000, fmt.Sprintf("%d", 100+i)))
	}
	server := klineServer(t, klines)
	defer server.Close()
	withServer(t, server)

	got, err := FetchKlines("BTCUSDT", "1m", 2500)
	if err != nil {
		t.Fatalf("FetchKlines: %v", err)
	}
	if len(got) != 2500 {
		t.Fatalf("len = %d, want 2500", len(got))
	}
	// oldest first
	if got[0].OpenTime != 1_000_000 || got[2499].OpenTime != 1_000_000+2499*60_000 {
		t.Fatalf("order wrong: first=%d last=%d", got[0].OpenTime, got[2499].OpenTime)
	}
}

func TestFetchKlinesRangePaginatesForward(t *testing.T) {
	var klines [][]any
	for i := int64(0); i < 2500; i++ {
		klines = append(klines, klineJSON(1_000_000+i*60_000, "100"))
	}
	server := klineServer(t, klines)
	defer server.Close()
	withServer(t, server)

	start := int64(1_000_000)
	end := int64(1_000_000 + 2499*60_000)
	got, err := FetchKlinesRange("BTCUSDT", "1m", start, end)
	if err != nil {
		t.Fatalf("FetchKlinesRange: %v", err)
	}
	if len(got) != 2500 {
		t.Fatalf("len = %d, want 2500", len(got))
	}
	if got[0].OpenTime != start || got[len(got)-1].OpenTime != end {
		t.Fatalf("bounds wrong: first=%d last=%d", got[0].OpenTime, got[len(got)-1].OpenTime)
	}
}

func TestFetchKlinesRangeRespectsEnd(t *testing.T) {
	var klines [][]any
	for i := int64(0); i < 10; i++ {
		klines = append(klines, klineJSON(1_000_000+i*60_000, "100"))
	}
	server := klineServer(t, klines)
	defer server.Close()
	withServer(t, server)

	got, err := FetchKlinesRange("BTCUSDT", "1m", 1_000_000, 1_000_000+2*60_000)
	if err != nil {
		t.Fatalf("FetchKlinesRange: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
}

func TestFetchPrice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"price": "123.45"})
	}))
	defer server.Close()
	withServer(t, server)

	price, err := FetchPrice("BTCUSDT")
	if err != nil {
		t.Fatalf("FetchPrice: %v", err)
	}
	if price != "123.45" {
		t.Fatalf("price = %q", price)
	}
}

func TestRawKlineUnmarshal(t *testing.T) {
	var k RawKline
	err := json.Unmarshal([]byte(`[1000,"1","2","0.5","1.5","10",2000,"0",0,"0","0","0"]`), &k)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if k.OpenTime != 1000 || k.Close != "1.5" || k.CloseTime != 2000 {
		t.Fatalf("got %+v", k)
	}

	var short RawKline
	if err := json.Unmarshal([]byte(`[1000,"1"]`), &short); err == nil {
		t.Fatal("expected error for short kline array")
	}
}

func TestSignedRequest(t *testing.T) {
	t.Setenv("BINANCE_API_KEY", "testkey")
	t.Setenv("BINANCE_SECRET_KEY", "testsecret")

	params := url.Values{}
	params.Set("symbol", "BTCUSDT")
	params.Set("side", "BUY")

	req, err := signedRequest(http.MethodPost, "/api/v3/order", params)
	if err != nil {
		t.Fatalf("signedRequest: %v", err)
	}
	if req.Header.Get("X-MBX-APIKEY") != "testkey" {
		t.Fatalf("api key header = %q", req.Header.Get("X-MBX-APIKEY"))
	}

	q := req.URL.Query()
	sig := q.Get("signature")
	if sig == "" {
		t.Fatal("missing signature")
	}
	if q.Get("timestamp") == "" || q.Get("recvWindow") != "5000" {
		t.Fatalf("missing timestamp/recvWindow: %v", q)
	}

	// recompute the expected signature over the query without the signature param
	q.Del("signature")
	mac := hmac.New(sha256.New, []byte("testsecret"))
	mac.Write([]byte(q.Encode()))
	want := hex.EncodeToString(mac.Sum(nil))
	if sig != want {
		t.Fatalf("signature mismatch: got %s want %s", sig, want)
	}
}

func TestSignedRequestRequiresEnv(t *testing.T) {
	t.Setenv("BINANCE_API_KEY", "")
	t.Setenv("BINANCE_SECRET_KEY", "")

	_, err := signedRequest(http.MethodPost, "/api/v3/order", url.Values{})
	if err == nil {
		t.Fatal("expected error when credentials are not set")
	}
}

func TestNormalizeSymbol(t *testing.T) {
	cases := map[string]string{
		"XRP/USDT":   "XRPUSDT",
		"xrp/usdt":   "XRPUSDT",
		"BTC-USDT":   "BTCUSDT",
		"BTCUSDT":    "BTCUSDT",
		"btc.usdt":   "BTCUSDT",
		" SOL/USDC ": "SOLUSDC",
	}
	for in, want := range cases {
		if got := NormalizeSymbol(in); got != want {
			t.Errorf("NormalizeSymbol(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPlaceTradeNormalizesSymbol(t *testing.T) {
	t.Setenv("BINANCE_API_KEY", "k")
	t.Setenv("BINANCE_SECRET_KEY", "s")

	var gotQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/order" || r.Method != http.MethodPost {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		gotQuery = r.URL.Query()
		_ = json.NewEncoder(w).Encode(map[string]string{"price": "1.23"})
	}))
	defer server.Close()
	withServer(t, server)

	price, err := PlaceTrade("xrp/usdt", engine.BUY, 0.01)
	if err != nil {
		t.Fatalf("PlaceTrade: %v", err)
	}
	if price != "1.23" {
		t.Fatalf("price = %q", price)
	}
	if gotQuery.Get("symbol") != "XRPUSDT" {
		t.Errorf("symbol = %q, want XRPUSDT", gotQuery.Get("symbol"))
	}
	if gotQuery.Get("side") != "BUY" || gotQuery.Get("type") != "MARKET" || gotQuery.Get("quantity") != "0.01" {
		t.Errorf("query = %v", gotQuery)
	}
	if gotQuery.Get("signature") == "" {
		t.Error("missing signature")
	}
}
