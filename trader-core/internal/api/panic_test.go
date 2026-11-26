package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"trader-core/internal/bot"
	"trader-core/internal/db/models"
	"trader-core/internal/dto"
	"trader-core/internal/engine"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

func setupPanicRouter(t *testing.T, b *bot.Bot) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	dispatcher := bot.NewDispatcher()
	mm := bot.NewMarketDataManager(nil, dispatcher, nil)
	go mm.Run(ctx)
	deadline := time.Now().Add(2 * time.Second)
	for !mm.IsRunning() && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if !mm.IsRunning() {
		t.Fatal("market manager did not start")
	}

	origRuntime := runtime
	origBots := activeBots
	runtime = &bot.Runtime{Dispatcher: dispatcher, MarketManager: mm}
	activeBots = map[string]*bot.Bot{b.ID: b}
	t.Cleanup(func() { runtime = origRuntime; activeBots = origBots })

	r := gin.New()
	r.POST("/panic", panicHandler)
	return r
}

func newRunningBot(t *testing.T, id string, withPosition bool) *bot.Bot {
	t.Helper()
	acc := engine.NewPaperAccount("10000", "0.001")
	pe := engine.NewPaperExecutor(acc)
	if withPosition {
		if _, err := pe.ExecuteFill(engine.Order{
			BotID: id, Symbol: "BTCUSDT", Signal: "BUY",
			Price: decimal.NewFromInt(100), Qty: decimal.NewFromInt(1),
		}); err != nil {
			t.Fatal(err)
		}
	}
	return &bot.Bot{
		BotConfig: bot.BotConfig{
			ID: id, Mode: bot.BotModePaper, Exchange: "binance",
			Base: "BTC", Quote: "USDT", Interval: "1h", Lookback: "200h0m0s",
		},
		Status:            bot.BotRunning,
		Engine:            pe,
		Logger:            bot.NewBotLogger(id),
		CandleCh:          make(chan models.Candle, 10),
		Candles:           []models.Candle{{Close: decimal.NewFromInt(110)}},
		FillBroadcaster:   &bot.Broadcaster[dto.FIllDTO]{},
		CandleBroadcaster: &bot.Broadcaster[models.Candle]{},
		TickBroadcaster:   &bot.Broadcaster[models.Candle]{},
	}
}

func TestPanicStopsRunningBots(t *testing.T) {
	b := newRunningBot(t, "b1", false)
	r := setupPanicRouter(t, b)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/panic", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var body struct {
		Stopped   []string `json:"stopped"`
		Flattened []string `json:"flattened"`
		Errors    []string `json:"errors"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Stopped) != 1 || body.Stopped[0] != "b1" {
		t.Fatalf("stopped = %v", body.Stopped)
	}
	if b.Status != bot.BotCreated {
		t.Fatalf("bot status = %s, want created", b.Status)
	}
	if len(body.Errors) != 0 {
		t.Fatalf("errors = %v", body.Errors)
	}
}

func TestPanicFlattenSellsPositions(t *testing.T) {
	b := newRunningBot(t, "b2", true)
	r := setupPanicRouter(t, b)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/panic", strings.NewReader(`{"flatten":true}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	var body struct {
		Stopped   []string `json:"stopped"`
		Flattened []string `json:"flattened"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Flattened) != 1 || body.Flattened[0] != "BTCUSDT" {
		t.Fatalf("flattened = %v", body.Flattened)
	}
	snap := b.Engine.(engine.AccountSnapshotter).Snapshot()
	if !snap.Positions["BTCUSDT"].IsZero() {
		t.Fatalf("position should be flat after panic flatten, got %s", snap.Positions["BTCUSDT"])
	}
}
