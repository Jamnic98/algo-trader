package engine

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"trader-core/internal/strategies"

	"github.com/shopspring/decimal"
)

// RiskLimits are per-bot risk controls. Percent fields are decimal strings
// ("5" = 5%); a "0" percent disables that kill-switch.
type RiskLimits struct {
	MaxPositionPct  string `json:"max_position_pct"`  // cap a single BUY notional as % of balance; default "100" (no cap)
	MaxDailyLossPct string `json:"max_daily_loss_pct"` // kill switch: equity drop vs day-start equity; "0" = off
	MaxDrawdownPct  string `json:"max_drawdown_pct"`   // kill switch: equity drawdown from running peak; "0" = off
	CooldownLosses  int    `json:"cooldown_losses"`    // consecutive losing SELLs before buys pause; 0 = off
	CooldownMinutes int    `json:"cooldown_minutes"`   // how long buys stay paused
}

// DefaultRiskLimits returns the safe defaults: only the position cap is set
// (100% = no cap), all kill-switches and the cooldown are off.
func DefaultRiskLimits() RiskLimits {
	return RiskLimits{MaxPositionPct: "100", MaxDailyLossPct: "0", MaxDrawdownPct: "0"}
}

// ParseRiskLimits unmarshals stored JSON over the defaults so missing fields
// fall back to safe values.
func ParseRiskLimits(raw []byte) (RiskLimits, error) {
	l := DefaultRiskLimits()
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &l); err != nil {
			return RiskLimits{}, fmt.Errorf("invalid risk config: %w", err)
		}
	}
	if l.MaxPositionPct == "" {
		l.MaxPositionPct = "100"
	}
	pct, err := decimal.NewFromString(l.MaxPositionPct)
	if err != nil || pct.LessThan(decimal.Zero) || pct.GreaterThan(decimal.NewFromInt(100)) {
		return RiskLimits{}, fmt.Errorf("max_position_pct must be a decimal in [0,100], got %q", l.MaxPositionPct)
	}
	for _, f := range []string{l.MaxDailyLossPct, l.MaxDrawdownPct} {
		if f == "" {
			continue
		}
		if p, err := decimal.NewFromString(f); err != nil || p.LessThan(decimal.Zero) {
			return RiskLimits{}, fmt.Errorf("percent fields must be non-negative decimals, got %q", f)
		}
	}
	if l.CooldownLosses < 0 || l.CooldownMinutes < 0 {
		return RiskLimits{}, fmt.Errorf("cooldown fields must be >= 0")
	}
	return l, nil
}

func parsePct(s string) decimal.Decimal {
	if s == "" {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(s)
	if err != nil {
		return decimal.Zero
	}
	return d
}

// RiskState is the per-running-bot risk tracker (in-memory; reset on attach).
type RiskState struct {
	mu             sync.Mutex
	limits         RiskLimits
	dayKey         string // UTC day of dayStartEquity
	dayStartEquity decimal.Decimal
	peakEquity     decimal.Decimal
	consecLosses   int
	cooldownUntil  time.Time
	tripped        string // non-empty once a kill-switch has fired
}

func NewRiskState(limits RiskLimits) *RiskState {
	return &RiskState{limits: limits}
}

// TrackEquity updates day-start/peak tracking and returns a kill reason if a
// kill-switch trips ("" otherwise). Once tripped it keeps returning the reason.
func (r *RiskState) TrackEquity(equity decimal.Decimal, now time.Time) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.tripped != "" {
		return r.tripped
	}

	day := now.UTC().Format("2006-01-02")
	if r.dayKey == "" || day != r.dayKey {
		r.dayKey = day
		r.dayStartEquity = equity
		r.consecLosses = 0
	}
	if equity.GreaterThan(r.peakEquity) {
		r.peakEquity = equity
	}

	hundred := decimal.NewFromInt(100)
	if d := parsePct(r.limits.MaxDailyLossPct); d.GreaterThan(decimal.Zero) && r.dayStartEquity.GreaterThan(decimal.Zero) {
		lossPct := r.dayStartEquity.Sub(equity).Div(r.dayStartEquity).Mul(hundred)
		if lossPct.GreaterThanOrEqual(d) {
			r.tripped = fmt.Sprintf("daily loss %.2f%% >= limit %.2f%%", lossPct.InexactFloat64(), d.InexactFloat64())
			return r.tripped
		}
	}
	if d := parsePct(r.limits.MaxDrawdownPct); d.GreaterThan(decimal.Zero) && r.peakEquity.GreaterThan(decimal.Zero) {
		ddPct := r.peakEquity.Sub(equity).Div(r.peakEquity).Mul(hundred)
		if ddPct.GreaterThanOrEqual(d) {
			r.tripped = fmt.Sprintf("drawdown %.2f%% >= limit %.2f%%", ddPct.InexactFloat64(), d.InexactFloat64())
			return r.tripped
		}
	}
	return ""
}

// Tripped reports the kill-switch reason ("" if not tripped).
func (r *RiskState) Tripped() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.tripped
}

// CapQty caps a BUY quantity so its notional stays within MaxPositionPct of
// the balance. Returns the (possibly reduced) qty and whether it was capped.
func (r *RiskState) CapQty(qty, price, balance decimal.Decimal) (decimal.Decimal, bool) {
	pct := parsePct(r.limits.MaxPositionPct)
	if pct.GreaterThanOrEqual(decimal.NewFromInt(100)) || price.LessThanOrEqual(decimal.Zero) || balance.LessThanOrEqual(decimal.Zero) {
		return qty, false
	}
	maxQty := balance.Mul(pct).Div(decimal.NewFromInt(100)).Div(price)
	if qty.GreaterThan(maxQty) {
		return maxQty, true
	}
	return qty, false
}

// NoteFill updates the consecutive-loss counter from a fill's realized pnl
// (SELL with pnl < 0 counts as a loss) and arms the cooldown when the limit
// is reached.
func (r *RiskState) NoteFill(signal strategies.Signal, pnl decimal.Decimal, now time.Time) {
	if r.limits.CooldownLosses <= 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if signal == strategies.Sell && pnl.LessThan(decimal.Zero) {
		r.consecLosses++
		if r.consecLosses >= r.limits.CooldownLosses {
			r.cooldownUntil = now.Add(time.Duration(r.limits.CooldownMinutes) * time.Minute)
			r.consecLosses = 0
		}
	} else if signal == strategies.Sell {
		r.consecLosses = 0
	}
}

// Blocked reports whether new BUYs are paused by the cooldown.
func (r *RiskState) Blocked(now time.Time) (string, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cooldownUntil.IsZero() || !now.Before(r.cooldownUntil) {
		return "", false
	}
	return fmt.Sprintf("cooldown until %s", r.cooldownUntil.UTC().Format(time.RFC3339)), true
}
