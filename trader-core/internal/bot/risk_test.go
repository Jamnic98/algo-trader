package bot

import (
	"testing"

	"trader-core/internal/db/models"
	"trader-core/internal/dto"
	"trader-core/internal/engine"

	"github.com/shopspring/decimal"
)

func TestBuildBotConfigRiskDefaults(t *testing.T) {
	cfg, err := BuildBotConfig("id1", CreateBotData{StrategySlug: "simple", Interval: "1h", MaxCandles: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Risk) == 0 {
		t.Fatal("risk should default to a stored JSON, not empty")
	}
	limits, err := cfg.RiskLimits()
	if err != nil {
		t.Fatal(err)
	}
	def := engine.DefaultRiskLimits()
	if limits != def {
		t.Fatalf("limits = %+v, want defaults %+v", limits, def)
	}
}

func TestBuildBotConfigRiskCustom(t *testing.T) {
	cfg, err := BuildBotConfig("id1", CreateBotData{
		StrategySlug: "simple", Interval: "1h", MaxCandles: 10,
		Risk: []byte(`{"max_daily_loss_pct":"5","max_position_pct":"25","cooldown_losses":3,"cooldown_minutes":60}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	limits, err := cfg.RiskLimits()
	if err != nil {
		t.Fatal(err)
	}
	if limits.MaxDailyLossPct != "5" || limits.MaxPositionPct != "25" || limits.CooldownLosses != 3 || limits.CooldownMinutes != 60 {
		t.Fatalf("limits = %+v", limits)
	}
}

func TestBuildBotConfigRiskInvalid(t *testing.T) {
	if _, err := BuildBotConfig("id1", CreateBotData{
		StrategySlug: "simple", Interval: "1h", MaxCandles: 10,
		Risk: []byte(`{"max_position_pct":"999"}`),
	}); err == nil {
		t.Fatal("expected error for out-of-range max_position_pct")
	}
}

func TestFlattenSellsOpenPosition(t *testing.T) {
	acc := engine.NewPaperAccount("10000", "0.001")
	pe := engine.NewPaperExecutor(acc)

	// bot holds 2 BTC bought at 100
	if _, err := pe.ExecuteFill(engine.Order{BotID: "b1", Symbol: "BTCUSDT", Signal: "BUY", Price: decimal.NewFromInt(100), Qty: decimal.NewFromInt(2)}); err != nil {
		t.Fatal(err)
	}

	b := &Bot{
		BotConfig:         BotConfig{ID: "b1", Exchange: "binance", Base: "BTC", Quote: "USDT", Interval: "1h"},
		Status:            BotRunning,
		Engine:            pe,
		Logger:            NewBotLogger("b1"),
		CandleCh:          make(chan models.Candle, 10),
		Candles:           []models.Candle{{Close: decimal.NewFromInt(110)}},
		FillBroadcaster:   &Broadcaster[dto.FIllDTO]{},
		CandleBroadcaster: &Broadcaster[models.Candle]{},
		TickBroadcaster:   &Broadcaster[models.Candle]{},
	}

	flattened, err := Flatten(b)
	if err != nil {
		t.Fatalf("Flatten: %v", err)
	}
	if len(flattened) != 1 || flattened[0] != "BTCUSDT" {
		t.Fatalf("flattened = %v", flattened)
	}
	snap := acc.Snapshot()
	if !snap.Positions["BTCUSDT"].IsZero() {
		t.Fatalf("position should be flat, got %s", snap.Positions["BTCUSDT"])
	}
	// 10000 - 200 - 0.2 (buy fee) + 220 - 0.22 (sell fee) = 10019.58
	want := decimal.NewFromFloat(10019.58)
	if !snap.Balance.Equal(want) {
		t.Fatalf("balance = %s, want %s", snap.Balance, want)
	}
}

func TestFlattenNoCandles(t *testing.T) {
	acc := engine.NewPaperAccount("10000", "0.001")
	b := &Bot{
		BotConfig:       BotConfig{ID: "b1"},
		Engine:          engine.NewPaperExecutor(acc),
		Logger:          NewBotLogger("b1"),
		FillBroadcaster: &Broadcaster[dto.FIllDTO]{},
	}
	if _, err := Flatten(b); err == nil {
		t.Fatal("expected error when no candles loaded")
	}
}
