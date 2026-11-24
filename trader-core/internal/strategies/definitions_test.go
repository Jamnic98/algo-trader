package strategies

import (
	"testing"

	"trader-core/internal/db/models"

	"github.com/shopspring/decimal"
)

// TestSharedSchemaDrivesRegistry ensures every primitive declared in
// shared/strategies.json is registered and served by Schemas().
func TestSharedSchemaDrivesRegistry(t *testing.T) {
	defs := Definitions()
	if len(defs.Primitives) == 0 {
		t.Fatal("no primitives in shared schema")
	}

	schemas := map[string]StrategySchema{}
	for _, s := range Schemas() {
		schemas[s.Name] = s
	}

	for _, def := range defs.Primitives {
		if _, err := Build(def.Name, []byte("{}")); err != nil {
			t.Errorf("Build(%q): %v", def.Name, err)
			continue
		}
		schema, ok := schemas[def.Name]
		if !ok {
			t.Errorf("schema missing for %q", def.Name)
			continue
		}
		if schema.DisplayName != def.DisplayName {
			t.Errorf("%q display name = %q, want %q", def.Name, schema.DisplayName, def.DisplayName)
		}
		if len(schema.Params) != len(def.Params) {
			t.Errorf("%q params = %d, want %d", def.Name, len(schema.Params), len(def.Params))
		}
	}
}

// TestSharedSchemaDefaults ensures missing params fall back to schema defaults.
func TestSharedSchemaDefaults(t *testing.T) {
	def := FindPrimitive("smart_dca")
	if def == nil {
		t.Fatal("smart_dca not in shared schema")
	}

	s, err := Build("smart_dca", []byte("{}"))
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	dca, ok := s.(*SmartDCAStrategy)
	if !ok {
		t.Fatalf("got %T", s)
	}
	if !dca.fixedSpend.Equal(decimal.NewFromInt(10)) {
		t.Errorf("fixedSpend = %s, want 10", dca.fixedSpend)
	}
	if !dca.takeProfit.Equal(decimal.NewFromFloat(1.05)) {
		t.Errorf("takeProfit = %s, want 1.05", dca.takeProfit)
	}
}

// TestSharedTestVectors executes the behavior vectors from
// shared/strategies.json against the Go implementations. The Python backtester
// runs the same vectors (research/tests/test_definitions.py) — if these fail
// on either side, the engines have drifted.
func TestSharedTestVectors(t *testing.T) {
	for _, def := range Definitions().Primitives {
		feeRate, err := decimal.NewFromString(def.Test.FeeRate)
		if err != nil {
			t.Fatalf("%s: bad fee_rate: %v", def.Name, err)
		}
		for ci, tc := range def.Test.Cases {
			if len(tc.Candles) != len(tc.Expect) {
				t.Fatalf("%s case %d: candles/expect length mismatch", def.Name, ci)
			}
			strat, err := Build(def.Name, []byte("{}"))
			if err != nil {
				t.Fatalf("%s: %v", def.Name, err)
			}
			for i, close := range tc.Candles {
				c := decimal.NewFromFloat(close)
				d := strat.OnCandle(CandleContext{
					Candle:  models.Candle{Close: c, Open: c, High: c, Low: c},
					FeeRate: feeRate,
				})
				if string(d.Signal) != tc.Expect[i] {
					t.Errorf("%s case %d candle %d (close=%v): signal = %s, want %s",
						def.Name, ci, i, close, d.Signal, tc.Expect[i])
				}
			}
		}
	}
}

// TestUnknownTemplateRejected guards against typos in the shared schema.
func TestUnknownTemplateRejected(t *testing.T) {
	if _, err := makeConstructor(PrimitiveDef{Name: "x", Template: "nope"})([]byte("{}")); err == nil {
		t.Fatal("expected error for unknown template")
	}
}
