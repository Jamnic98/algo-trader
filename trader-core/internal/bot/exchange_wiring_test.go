package bot

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"trader-core/internal/db/models"
	"trader-core/internal/engine"
	"trader-core/internal/exchange"

	"github.com/shopspring/decimal"
)

// fakeStreamBroker is an exchange.Broker whose StreamKlines blocks until ctx
// is cancelled and records the onKline callback so tests can push candles.
type fakeStreamBroker struct {
	mu        sync.Mutex
	klines    []exchange.Kline
	streams   int
	onKline   func(exchange.Kline)
	cancelled atomic.Bool
}

func (f *fakeStreamBroker) Name() string { return "fakevenue" }
func (f *fakeStreamBroker) Klines(symbol, interval string, startMs, endMs int64, limit int) ([]exchange.Kline, error) {
	return f.klines, nil
}
func (f *fakeStreamBroker) StreamKlines(ctx context.Context, symbol, interval string, onKline func(exchange.Kline)) error {
	f.mu.Lock()
	f.streams++
	f.onKline = onKline
	f.mu.Unlock()
	<-ctx.Done()
	f.cancelled.Store(true)
	return ctx.Err()
}
func (f *fakeStreamBroker) Balances() (map[string]decimal.Decimal, error) {
	return map[string]decimal.Decimal{
		"USD": decimal.NewFromInt(100),
		"BTC": decimal.NewFromFloat(0.5),
	}, nil
}
func (f *fakeStreamBroker) Price(symbol string) (decimal.Decimal, error) {
	return decimal.NewFromInt(42), nil
}
func (f *fakeStreamBroker) PlaceMarketOrder(symbol string, side string, qty decimal.Decimal, clientOrderID string) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	return decimal.NewFromInt(1), qty, decimal.Zero, nil
}
func (f *fakeStreamBroker) Filters(symbol string) (engine.SymbolFilters, error) {
	return engine.SymbolFilters{}, nil
}

func (f *fakeStreamBroker) streamCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.streams
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func TestBuildBotConfigDefaultsExchange(t *testing.T) {
	cfg, err := BuildBotConfig("id1", CreateBotData{StrategySlug: "simple", Interval: "1h", MaxCandles: 10})
	if err != nil {
		t.Fatalf("BuildBotConfig: %v", err)
	}
	if cfg.Exchange != "binance" {
		t.Fatalf("default exchange = %q, want binance", cfg.Exchange)
	}
}

func TestBotSymbolPerExchange(t *testing.T) {
	cases := []struct {
		exchange, base, quote, assetType, want string
	}{
		{"binance", "btc", "usdt", "crypto", "BTCUSDT"},
		{"binance", "xrp/usdt", "", "crypto", "XRPUSDT"},
		{"alpaca", "BTC", "USD", "crypto", "BTC/USD"},
		{"alpaca", "aapl", "usd", "stocks", "AAPL"},
	}
	for _, c := range cases {
		b := &Bot{BotConfig: BotConfig{Exchange: c.exchange, Base: c.base, Quote: c.quote, AssetType: c.assetType}}
		if got := b.Symbol(); got != c.want {
			t.Errorf("Symbol(%s %s/%s %s) = %q, want %q", c.exchange, c.base, c.quote, c.assetType, got, c.want)
		}
	}
}

func TestMarketManagerBrokerStream(t *testing.T) {
	br := &fakeStreamBroker{}
	m := NewMarketDataManager(nil, NewDispatcher(), map[string]exchange.Broker{"fakevenue": br})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.ctx = ctx

	interval, err := engine.ParseInterval("1m")
	if err != nil {
		t.Fatal(err)
	}
	b := &Bot{
		BotConfig: BotConfig{ID: "b1", Exchange: "fakevenue", Base: "BTC", Quote: "USD", AssetType: "crypto"},
		Logger:    NewBotLogger("b1"),
		CandleCh:  make(chan models.Candle, 10),
	}
	m.dispatcher.Subscribe(b.Symbol(), interval, b)

	if err := m.Subscribe("fakevenue", b.Symbol(), interval); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	waitFor(t, "stream to start", func() bool { return br.streamCount() == 1 })

	// push a closed kline through the broker callback; the bot must receive it
	br.mu.Lock()
	onKline := br.onKline
	br.mu.Unlock()
	onKline(exchange.Kline{OpenTime: 1, CloseTime: 2, Close: decimal.NewFromInt(100), Open: decimal.NewFromInt(99), High: decimal.NewFromInt(101), Low: decimal.NewFromInt(98), Volume: decimal.NewFromInt(5)})
	select {
	case c := <-b.CandleCh:
		if c.Close.Cmp(decimal.NewFromInt(100)) != 0 {
			t.Fatalf("candle close = %s, want 100", c.Close.String())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("bot did not receive candle from broker stream")
	}

	// second subscriber joins the same stream (no new goroutine)
	if err := m.Subscribe("fakevenue", b.Symbol(), interval); err != nil {
		t.Fatalf("Subscribe 2: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	if br.streamCount() != 1 {
		t.Fatalf("streams = %d, want 1 (shared)", br.streamCount())
	}

	// unsubscribing both cancels the stream
	m.Unsubscribe("fakevenue", b.Symbol(), interval)
	m.Unsubscribe("fakevenue", b.Symbol(), interval)
	waitFor(t, "stream cancel", func() bool { return br.cancelled.Load() })

	// unknown venue must error, not panic
	if err := m.Subscribe("nope", "X", interval); err == nil {
		t.Fatal("expected error for unknown exchange")
	}
}

func TestMarketManagerFetchCandlesBroker(t *testing.T) {
	now := time.Now().UnixMilli()
	br := &fakeStreamBroker{klines: []exchange.Kline{
		{OpenTime: now - 120_000, CloseTime: now - 60_001, Close: decimal.NewFromInt(1)},
		{OpenTime: now - 60_000, CloseTime: now + 60_000, Close: decimal.NewFromInt(2)}, // still open
	}}
	m := NewMarketDataManager(nil, NewDispatcher(), map[string]exchange.Broker{"fakevenue": br})
	interval, _ := engine.ParseInterval("1m")

	cs, err := m.FetchCandles("fakevenue", "BTC/USD", interval, 10)
	if err != nil {
		t.Fatalf("FetchCandles: %v", err)
	}
	if len(cs) != 1 || cs[0].Close.Cmp(decimal.NewFromInt(1)) != 0 {
		t.Fatalf("candles = %+v, want only the closed one", cs)
	}

	if _, err := m.FetchCandles("nope", "X", interval, 10); err == nil {
		t.Fatal("expected error for unknown exchange")
	}
}

func TestMarketManagerExchangeReady(t *testing.T) {
	m := NewMarketDataManager(nil, NewDispatcher(), map[string]exchange.Broker{"fakevenue": &fakeStreamBroker{}})
	if !m.ExchangeReady("fakevenue") {
		t.Fatal("fakevenue should be ready")
	}
	if m.ExchangeReady("binance") {
		t.Fatal("binance not ready without a live client")
	}
	if m.ExchangeReady("nope") {
		t.Fatal("unknown venue should not be ready")
	}
}

func TestNewBotLiveUsesExchangeBroker(t *testing.T) {
	br := &fakeStreamBroker{}
	live := engine.NewLiveAccount("0", "0.001")
	rt := &Runtime{
		Accounts:   engine.NewAccountManager(nil, live),
		BotFactory: &BotFactory{LiveAccount: live},
		Brokers:    map[string]exchange.Broker{"fakevenue": br},
	}

	cfg := BotConfig{ID: "b1", Mode: BotModeLive, Exchange: "fakevenue", Base: "BTC", Quote: "USD", AssetType: "crypto"}
	b, err := rt.newBot(cfg, nil)
	if err != nil {
		t.Fatalf("newBot: %v", err)
	}
	le, ok := b.Engine.(*engine.LiveExecutor)
	if !ok {
		t.Fatalf("engine = %T, want *engine.LiveExecutor", b.Engine)
	}
	if le.Placer != exchange.Broker(br) {
		t.Fatalf("placer = %T, want the fakevenue broker", le.Placer)
	}
	// live ledger synced from the venue balances with the bot's quote asset
	snap := rt.Accounts.Live().Snapshot()
	if snap.Balance.Cmp(decimal.NewFromInt(100)) != 0 {
		t.Fatalf("live balance = %s, want 100 USD", snap.Balance.String())
	}
	if snap.Positions["BTC"].Cmp(decimal.NewFromFloat(0.5)) != 0 {
		t.Fatalf("live positions = %v", snap.Positions)
	}
}

func TestNewBotLiveUnknownExchange(t *testing.T) {
	rt := &Runtime{
		Accounts:   engine.NewAccountManager(nil, engine.NewLiveAccount("0", "0.001")),
		BotFactory: &BotFactory{},
		Brokers:    map[string]exchange.Broker{},
	}
	cfg := BotConfig{ID: "b1", Mode: BotModeLive, Exchange: "nope", Quote: "USD"}
	if _, err := rt.newBot(cfg, nil); err == nil {
		t.Fatal("expected error for unknown exchange")
	}
}


// fakeTickBroker additionally implements exchange.TickStreamer.
type fakeTickBroker struct {
	fakeStreamBroker
	onKline2 func(exchange.Kline)
	onTick   func(exchange.Kline)
}

func (f *fakeTickBroker) StreamKlineEvents(ctx context.Context, symbol, interval string, onKline, onTick func(exchange.Kline)) error {
	f.mu.Lock()
	f.streams++
	f.onKline2 = onKline
	f.onTick = onTick
	f.mu.Unlock()
	<-ctx.Done()
	f.cancelled.Store(true)
	return ctx.Err()
}

// TestMarketManagerBrokerTicks: brokers implementing exchange.TickStreamer get
// in-progress bars dispatched as ticks (the candles chart's live line).
func TestMarketManagerBrokerTicks(t *testing.T) {
	br := &fakeTickBroker{}
	m := NewMarketDataManager(nil, NewDispatcher(), map[string]exchange.Broker{"tickvenue": br})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.ctx = ctx

	interval, _ := engine.ParseInterval("1m")
	b := &Bot{
		BotConfig:       BotConfig{ID: "b1", Exchange: "tickvenue", Base: "BTC", Quote: "USD", AssetType: "crypto"},
		Logger:          NewBotLogger("b1"),
		CandleCh:        make(chan models.Candle, 10),
		TickBroadcaster: &Broadcaster[models.Candle]{},
	}
	tickCh := b.TickBroadcaster.Subscribe()
	defer b.TickBroadcaster.Unsubscribe(tickCh)
	m.dispatcher.Subscribe(b.Symbol(), interval, b)

	if err := m.Subscribe("tickvenue", b.Symbol(), interval); err != nil {
		t.Fatalf("Subscribe: %v", err)
	}
	waitFor(t, "tick stream to start", func() bool { return br.streamCount() == 1 })

	k := exchange.Kline{OpenTime: 1, CloseTime: 2, Close: decimal.NewFromInt(100)}
	br.mu.Lock()
	onKline, onTick := br.onKline2, br.onTick
	br.mu.Unlock()

	onTick(k) // in-progress bar -> tick only
	select {
	case tick := <-tickCh:
		if tick.Close.Cmp(decimal.NewFromInt(100)) != 0 {
			t.Fatalf("tick close = %s", tick.Close.String())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no tick dispatched")
	}
	select {
	case <-b.CandleCh:
		t.Fatal("tick must not be dispatched as a closed candle")
	default:
	}

	onKline(k) // closed bar -> candle
	select {
	case c := <-b.CandleCh:
		if c.Close.Cmp(decimal.NewFromInt(100)) != 0 {
			t.Fatalf("candle close = %s", c.Close.String())
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no candle dispatched")
	}
}
