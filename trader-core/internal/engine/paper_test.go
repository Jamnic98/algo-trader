package engine

import (
	"testing"

	"trader-core/internal/strategies"

	"github.com/shopspring/decimal"
)

func TestPaperExecutorBuySellRoundTrip(t *testing.T) {
	acc := NewPaperAccount("1000", "0.001")
	pe := NewPaperExecutor(acc)

	fill, err := pe.ExecuteFill(Order{
		BotID:  "b1",
		Symbol: "BTCUSDT",
		Signal: strategies.Buy,
		Price:  decimal.NewFromInt(100),
		Qty:    decimal.NewFromInt(2),
	})
	if err != nil {
		t.Fatalf("buy: %v", err)
	}
	if !fill.Fee.Equal(decimal.NewFromFloat(0.2)) {
		t.Fatalf("fee = %s, want 0.2", fill.Fee)
	}
	if !acc.balance.Equal(decimal.NewFromFloat(799.8)) {
		t.Fatalf("balance = %s, want 799.8", acc.balance)
	}
	if !acc.Positions["BTCUSDT"].Equal(decimal.NewFromInt(2)) {
		t.Fatalf("position = %s, want 2", acc.Positions["BTCUSDT"])
	}

	_, err = pe.ExecuteFill(Order{
		BotID:  "b1",
		Symbol: "BTCUSDT",
		Signal: strategies.Sell,
		Price:  decimal.NewFromInt(110),
		Qty:    decimal.NewFromInt(2),
	})
	if err != nil {
		t.Fatalf("sell: %v", err)
	}
	if !acc.balance.Equal(decimal.NewFromFloat(1019.58)) {
		t.Fatalf("balance = %s, want 1019.58", acc.balance)
	}
	if !acc.Positions["BTCUSDT"].IsZero() {
		t.Fatalf("position = %s, want 0", acc.Positions["BTCUSDT"])
	}
}

func TestPaperExecutorInsufficientCash(t *testing.T) {
	acc := NewPaperAccount("100", "0.001")
	pe := NewPaperExecutor(acc)

	_, err := pe.ExecuteFill(Order{
		Symbol: "BTCUSDT",
		Signal: strategies.Buy,
		Price:  decimal.NewFromInt(100),
		Qty:    decimal.NewFromInt(2), // costs 200.2 > 100
	})
	if err == nil {
		t.Fatal("expected not enough cash error")
	}
}

func TestPaperExecutorInsufficientAsset(t *testing.T) {
	acc := NewPaperAccount("1000", "0.001")
	pe := NewPaperExecutor(acc)

	_, err := pe.ExecuteFill(Order{
		Symbol: "BTCUSDT",
		Signal: strategies.Sell,
		Price:  decimal.NewFromInt(100),
		Qty:    decimal.NewFromInt(1),
	})
	if err == nil {
		t.Fatal("expected not enough asset error")
	}
}

func TestPaperExecutorInvalidSide(t *testing.T) {
	acc := NewPaperAccount("1000", "0.001")
	pe := NewPaperExecutor(acc)

	_, err := pe.ExecuteFill(Order{
		Symbol: "BTCUSDT",
		Signal: strategies.Hold,
		Price:  decimal.NewFromInt(100),
		Qty:    decimal.NewFromInt(1),
	})
	if err == nil {
		t.Fatal("expected invalid side error")
	}
}

func TestPaperAccountSnapshotIsACopy(t *testing.T) {
	acc := NewPaperAccount("1000", "0.001")
	acc.Positions["BTCUSDT"] = decimal.NewFromInt(1)

	snap := acc.Snapshot()
	snap.Positions["BTCUSDT"] = decimal.NewFromInt(99)
	snap.Balance = decimal.NewFromInt(0)

	if !acc.Positions["BTCUSDT"].Equal(decimal.NewFromInt(1)) {
		t.Fatal("snapshot leaked into account positions")
	}
	if !acc.balance.Equal(decimal.NewFromInt(1000)) {
		t.Fatal("snapshot leaked into account balance")
	}
}
