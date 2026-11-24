package engine

import (
	"errors"
	"testing"

	"trader-core/internal/strategies"

	"github.com/shopspring/decimal"
)

type fakePlacer struct {
	avgPrice decimal.Decimal
	qty      decimal.Decimal
	fee      decimal.Decimal
	err      error

	gotSymbol string
	gotSide   string
	gotQty    decimal.Decimal
}

func (f *fakePlacer) PlaceMarketOrder(symbol string, side string, qty decimal.Decimal) (decimal.Decimal, decimal.Decimal, decimal.Decimal, error) {
	f.gotSymbol, f.gotSide, f.gotQty = symbol, side, qty
	if f.err != nil {
		return decimal.Zero, decimal.Zero, decimal.Zero, f.err
	}
	return f.avgPrice, f.qty, f.fee, nil
}

func liveOrder(side strategies.Signal, qty int64) Order {
	return Order{BotID: "b1", Symbol: "BTCUSDT", Signal: side, Price: decimal.NewFromInt(100), Qty: decimal.NewFromInt(qty)}
}

func TestLiveExecutorPlacesRealOrderAndRecordsFill(t *testing.T) {
	acc := NewLiveAccount("1000", "0.001")
	placer := &fakePlacer{avgPrice: decimal.NewFromInt(100), qty: decimal.NewFromInt(2), fee: decimal.NewFromFloat(0.2)}
	le := NewLiveExecutor(acc, placer)

	fill, err := le.ExecuteFill(liveOrder(strategies.Buy, 2))
	if err != nil {
		t.Fatalf("ExecuteFill: %v", err)
	}
	if placer.gotSymbol != "BTCUSDT" || placer.gotSide != "BUY" || !placer.gotQty.Equal(decimal.NewFromInt(2)) {
		t.Fatalf("placer got %s %s %s", placer.gotSymbol, placer.gotSide, placer.gotQty)
	}
	if !fill.Price.Equal(decimal.NewFromInt(100)) || !fill.Fee.Equal(decimal.NewFromFloat(0.2)) {
		t.Fatalf("fill = %+v", fill)
	}
	// ledger updated: 1000 - 200 - 0.2
	if !acc.Snapshot().Balance.Equal(decimal.NewFromFloat(799.8)) {
		t.Fatalf("balance = %s", acc.Snapshot().Balance)
	}
	if !acc.Snapshot().Positions["BTCUSDT"].Equal(decimal.NewFromInt(2)) {
		t.Fatalf("position = %s", acc.Snapshot().Positions["BTCUSDT"])
	}

	// sell 2 @ 110
	placer.avgPrice = decimal.NewFromInt(110)
	placer.qty = decimal.NewFromInt(2)
	placer.fee = decimal.NewFromFloat(0.22)
	fill, err = le.ExecuteFill(liveOrder(strategies.Sell, 2))
	if err != nil {
		t.Fatalf("sell: %v", err)
	}
	if fill.Signal != strategies.Sell {
		t.Fatalf("fill signal = %v", fill.Signal)
	}
	if !acc.Snapshot().Balance.Equal(decimal.NewFromFloat(1019.58)) {
		t.Fatalf("balance = %s, want 1019.58", acc.Snapshot().Balance)
	}
}

func TestLiveExecutorPropagatesPlacerError(t *testing.T) {
	acc := NewLiveAccount("1000", "0.001")
	placer := &fakePlacer{err: errors.New("binance down")}
	le := NewLiveExecutor(acc, placer)

	if _, err := le.ExecuteFill(liveOrder(strategies.Buy, 1)); err == nil {
		t.Fatal("expected placer error to propagate")
	}
	// ledger untouched
	if !acc.Snapshot().Balance.Equal(decimal.NewFromInt(1000)) {
		t.Fatal("balance changed despite failed order")
	}
}

func TestLiveExecutorRequiresPlacer(t *testing.T) {
	le := NewLiveExecutor(NewLiveAccount("1000", "0.001"), nil)
	if _, err := le.ExecuteFill(liveOrder(strategies.Buy, 1)); err == nil {
		t.Fatal("expected error when no placer configured")
	}
}

func TestLiveExecutorFeeFallback(t *testing.T) {
	acc := NewLiveAccount("1000", "0.001")
	placer := &fakePlacer{avgPrice: decimal.NewFromInt(100), qty: decimal.NewFromInt(1), fee: decimal.Zero}
	le := NewLiveExecutor(acc, placer)

	fill, err := le.ExecuteFill(liveOrder(strategies.Buy, 1))
	if err != nil {
		t.Fatalf("ExecuteFill: %v", err)
	}
	// fee falls back to notional * account fee rate = 100 * 1 * 0.001 = 0.1
	if !fill.Fee.Equal(decimal.NewFromFloat(0.1)) {
		t.Fatalf("fee = %s, want 0.1", fill.Fee)
	}
}
