package engine

import (
	"testing"
	"time"

	"trader-core/internal/strategies"

	"github.com/shopspring/decimal"
)

func d(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func TestParseRiskLimitsDefaults(t *testing.T) {
	l, err := ParseRiskLimits(nil)
	if err != nil {
		t.Fatal(err)
	}
	if l.MaxPositionPct != "100" || l.MaxDailyLossPct != "0" || l.MaxDrawdownPct != "0" || l.CooldownLosses != 0 {
		t.Fatalf("defaults = %+v", l)
	}

	// partial JSON keeps defaults for missing fields
	l, err = ParseRiskLimits([]byte(`{"max_daily_loss_pct":"5"}`))
	if err != nil {
		t.Fatal(err)
	}
	if l.MaxDailyLossPct != "5" || l.MaxPositionPct != "100" {
		t.Fatalf("merged = %+v", l)
	}

	// invalid values rejected
	if _, err := ParseRiskLimits([]byte(`{"max_position_pct":"150"}`)); err == nil {
		t.Fatal("expected error for max_position_pct=150")
	}
	if _, err := ParseRiskLimits([]byte(`{"max_daily_loss_pct":"abc"}`)); err == nil {
		t.Fatal("expected error for non-numeric pct")
	}
	if _, err := ParseRiskLimits([]byte(`{invalid`)); err == nil {
		t.Fatal("expected error for bad json")
	}
}

func TestTrackEquityDailyLossKillSwitch(t *testing.T) {
	limits := RiskLimits{MaxPositionPct: "100", MaxDailyLossPct: "5"}
	r := NewRiskState(limits)
	now := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	if kill := r.TrackEquity(d("10000"), now); kill != "" {
		t.Fatalf("unexpected kill: %s", kill)
	}
	// -4% -> fine
	if kill := r.TrackEquity(d("9600"), now); kill != "" {
		t.Fatalf("unexpected kill at -4%%: %s", kill)
	}
	// -6% -> trips
	kill := r.TrackEquity(d("9400"), now)
	if kill == "" {
		t.Fatal("expected daily-loss kill at -6% with 5% limit")
	}
	// stays tripped
	if kill2 := r.TrackEquity(d("9900"), now); kill2 == "" {
		t.Fatal("kill switch must stay tripped")
	}

	// next day resets the day-start baseline (but tripped state persists per run)
	r2 := NewRiskState(limits)
	nextDay := now.Add(24 * time.Hour)
	r2.TrackEquity(d("10000"), now)
	if kill := r2.TrackEquity(d("9500"), nextDay); kill != "" {
		t.Fatalf("new day should reset baseline, got kill: %s", kill)
	}
}

func TestTrackEquityDrawdownKillSwitch(t *testing.T) {
	limits := RiskLimits{MaxPositionPct: "100", MaxDrawdownPct: "10"}
	r := NewRiskState(limits)
	now := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	r.TrackEquity(d("10000"), now)
	r.TrackEquity(d("12000"), now) // peak 12000
	if kill := r.TrackEquity(d("10900"), now); kill != "" {
		t.Fatalf("-9.2%% from peak should not trip 10%% limit: %s", kill)
	}
	if kill := r.TrackEquity(d("10700"), now); kill == "" {
		t.Fatal("expected drawdown kill at -10.8% from peak with 10% limit")
	}
}

func TestTrackEquityDisabledByDefault(t *testing.T) {
	r := NewRiskState(DefaultRiskLimits())
	now := time.Now()
	r.TrackEquity(d("10000"), now)
	if kill := r.TrackEquity(d("1"), now); kill != "" {
		t.Fatalf("default limits must not trip: %s", kill)
	}
}

func TestCapQty(t *testing.T) {
	r := NewRiskState(RiskLimits{MaxPositionPct: "25"})
	// balance 1000, price 100 -> max qty = 1000*0.25/100 = 2.5
	q, capped := r.CapQty(d("10"), d("100"), d("1000"))
	if !capped || !q.Equal(d("2.5")) {
		t.Fatalf("q=%s capped=%v, want 2.5 true", q, capped)
	}
	// under the cap -> unchanged
	q, capped = r.CapQty(d("1"), d("100"), d("1000"))
	if capped || !q.Equal(d("1")) {
		t.Fatalf("q=%s capped=%v, want 1 false", q, capped)
	}
	// default 100% -> never caps
	r2 := NewRiskState(DefaultRiskLimits())
	if _, capped := r2.CapQty(d("1000"), d("1"), d("1")); capped {
		t.Fatal("default limits should not cap")
	}
}

func TestCooldownAfterLosses(t *testing.T) {
	limits := RiskLimits{MaxPositionPct: "100", CooldownLosses: 2, CooldownMinutes: 30}
	r := NewRiskState(limits)
	now := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)

	if _, blocked := r.Blocked(now); blocked {
		t.Fatal("should not be blocked initially")
	}
	r.NoteFill(strategies.Sell, d("-5"), now) // loss 1
	if _, blocked := r.Blocked(now); blocked {
		t.Fatal("one loss should not trigger cooldown at limit 2")
	}
	r.NoteFill(strategies.Sell, d("-3"), now) // loss 2 -> cooldown armed
	if reason, blocked := r.Blocked(now); !blocked {
		t.Fatal("expected cooldown after 2 consecutive losses")
	} else if reason == "" {
		t.Fatal("expected a reason string")
	}
	// still blocked 29 min later
	if _, blocked := r.Blocked(now.Add(29*time.Minute)); !blocked {
		t.Fatal("should still be blocked at 29m")
	}
	// unblocked after 30m
	later := now.Add(31 * time.Minute)
	if _, blocked := r.Blocked(later); blocked {
		t.Fatal("should be unblocked after cooldown")
	}
	// after cooldown: a winning sell resets the loss counter, one loss is not enough
	r.NoteFill(strategies.Sell, d("10"), later)
	r.NoteFill(strategies.Sell, d("-1"), later)
	if _, blocked := r.Blocked(later); blocked {
		t.Fatal("win should reset the loss counter")
	}
}
