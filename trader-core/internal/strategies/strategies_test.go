package strategies

import (
	"testing"

	"trader-core/internal/db/models"

	"github.com/shopspring/decimal"
)

func candle(close int64) models.Candle {
	c := decimal.NewFromInt(close)
	return models.Candle{Open: c, High: c, Low: c, Close: c, Volume: decimal.NewFromInt(1)}
}

func ctx(c models.Candle) CandleContext {
	return CandleContext{Candle: c, FeeRate: decimal.NewFromFloat(0.001)}
}

func TestRegistryBuildAndSchemas(t *testing.T) {
	if _, err := Build("simple", []byte("{}")); err != nil {
		t.Fatalf("build simple: %v", err)
	}
	if _, err := Build("simple_dca", []byte("{}")); err != nil {
		t.Fatalf("build simple_dca: %v", err)
	}
	if _, err := Build("does_not_exist", []byte("{}")); err == nil {
		t.Fatal("expected error for unknown strategy")
	}

	names := map[string]bool{}
	for _, s := range Schemas() {
		names[s.Name] = true
	}
	for _, want := range []string{"simple", "simple_dca", "smart_dca"} {
		if !names[want] {
			t.Fatalf("schemas missing %q (have %v)", want, names)
		}
	}
}

func TestSimpleStrategyBuyLowSellHigh(t *testing.T) {
	s := NewSimpleStrategy(SimpleParams{})

	if d := s.OnCandle(ctx(candle(100))); d.Signal != Hold {
		t.Fatalf("first candle: %v", d.Signal)
	}
	if d := s.OnCandle(ctx(candle(90))); d.Signal != Buy {
		t.Fatalf("down candle: %v", d.Signal)
	}
	if d := s.OnCandle(ctx(candle(95))); d.Signal != Sell {
		t.Fatalf("up candle: %v (want SELL, move 5 > round-trip fee)", d.Signal)
	}
	if d := s.OnCandle(ctx(candle(94))); d.Signal != Buy {
		t.Fatalf("down again: %v", d.Signal)
	}
}

func TestSimpleStrategyFeeGate(t *testing.T) {
	s := NewSimpleStrategy(SimpleParams{})
	s.OnCandle(ctx(candle(100)))
	s.OnCandle(ctx(candle(90))) // buy @90

	// up move of 0.05 on price 90 < round-trip fee 90*0.002 = 0.18 -> hold
	d := s.OnCandle(ctx(candle(90))) // equal close -> hold anyway
	_ = d
	s2 := NewSimpleStrategy(SimpleParams{})
	s2.OnCandle(ctx(candle(100)))
	s2.OnCandle(ctx(candle(90)))
	d2 := s2.OnCandle(CandleContext{
		Candle:  models.Candle{Close: decimal.NewFromFloat(90.05)},
		FeeRate: decimal.NewFromFloat(0.001),
	})
	if d2.Signal != Hold {
		t.Fatalf("small move: %v, want HOLD", d2.Signal)
	}
	if d2.Reason == "" {
		t.Fatal("expected fee gate reason")
	}
	if !s2.hasPosition {
		t.Fatal("should still be holding after fee-gated sell")
	}
}

func TestSimpleDCAStrategyAlwaysBuys(t *testing.T) {
	s := NewSimpleDCAStrategy(SimpleDCAParams{})
	for _, p := range []int64{100, 90, 95} {
		if d := s.OnCandle(ctx(candle(p))); d.Signal != Buy {
			t.Fatalf("price %d: %v, want BUY", p, d.Signal)
		}
	}
}

func TestSmartDCAStrategy(t *testing.T) {
	s := NewSmartDCAStrategy(SmartDCAParams{FixedSpend: "10.00"})

	d := s.OnCandle(ctx(candle(100)))
	if d.Signal != Buy || d.Quantity == nil || !d.Quantity.Equal(decimal.NewFromFloat(0.1)) {
		t.Fatalf("first: %+v, want BUY qty 0.1", d)
	}

	// 102 < avg entry 100 * 1.05 -> keep DCA-ing
	d = s.OnCandle(ctx(candle(102)))
	if d.Signal != Buy {
		t.Fatalf("below take-profit: %v, want BUY", d.Signal)
	}

	// reset and check sell above 105
	s2 := NewSmartDCAStrategy(SmartDCAParams{FixedSpend: "10.00"})
	s2.OnCandle(ctx(candle(100)))
	d2 := s2.OnCandle(ctx(candle(106)))
	if d2.Signal != Sell || d2.Quantity == nil || !d2.Quantity.Equal(decimal.NewFromFloat(0.1)) {
		t.Fatalf("above take-profit: %+v, want SELL qty 0.1", d2)
	}
}

type fakeStrategy struct{ d Decision }

func (f fakeStrategy) OnCandle(CandleContext) Decision { return f.d }

func TestWeightedComposer(t *testing.T) {
	c := candle(100)

	// both buy -> 100% >= 0.6 -> BUY
	comp := &WeightedComposer{Threshold: 0.6, Nodes: []ComposerNode{
		{Strategy: fakeStrategy{Decision{Signal: Buy}}, Weight: 1},
		{Strategy: fakeStrategy{Decision{Signal: Buy}}, Weight: 1},
	}}
	if d := comp.OnCandle(ctx(c)); d.Signal != Buy {
		t.Fatalf("both buy: %v", d.Signal)
	}

	// one buy one hold -> 50% < 60% -> HOLD
	comp.Nodes[1] = ComposerNode{Strategy: fakeStrategy{Decision{Signal: Hold}}, Weight: 1}
	d := comp.OnCandle(ctx(c))
	if d.Signal != Hold || d.Reason == "" {
		t.Fatalf("split vote: %+v, want HOLD with reason", d)
	}

	// zero total weight -> HOLD
	comp.Nodes = []ComposerNode{{Strategy: fakeStrategy{Decision{Signal: Buy}}, Weight: 0}}
	if d := comp.OnCandle(ctx(c)); d.Signal != Hold {
		t.Fatalf("zero weight: %v", d.Signal)
	}

	// sell side wins
	comp.Nodes = []ComposerNode{
		{Strategy: fakeStrategy{Decision{Signal: Sell}}, Weight: 2},
		{Strategy: fakeStrategy{Decision{Signal: Buy}}, Weight: 1},
	}
	if d := comp.OnCandle(ctx(c)); d.Signal != Sell {
		t.Fatalf("sell majority: %v", d.Signal)
	}
}
