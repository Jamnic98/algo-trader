package engine

import (
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// OrderPlacer places real orders on an exchange. A exchange.Broker satisfies
// this interface directly (side is "BUY"/"SELL"). clientOrderID is sent to the
// venue so a retried order can't double-fill.
type OrderPlacer interface {
	PlaceMarketOrder(symbol string, side string, qty decimal.Decimal, clientOrderID string) (avgPrice, executedQty, fee decimal.Decimal, err error)
}

// LiveAccount is the local ledger for live trading: initialized from the
// exchange's real balances, then updated from each executed fill.
type LiveAccount struct {
	mu        sync.RWMutex
	balance   decimal.Decimal            // quote currency (e.g. USDT) available
	Fee       decimal.Decimal            // fallback fee rate if the venue doesn't report one
	Positions map[string]decimal.Decimal // symbol -> qty
	costBasis map[string]decimal.Decimal // symbol -> quote spent (avg entry base)
}

func NewLiveAccount(startBalance, fee string) *LiveAccount {
	return &LiveAccount{
		balance:   decimal.RequireFromString(startBalance),
		Fee:       decimal.RequireFromString(fee),
		Positions: make(map[string]decimal.Decimal),
		costBasis: make(map[string]decimal.Decimal),
	}
}

// SetBalance sets the available quote-currency balance (e.g. from FetchAccount).
func (a *LiveAccount) SetBalance(amount decimal.Decimal) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.balance = amount
}

// LiveExecutor implements ExecutionEngine by placing real orders through an
// OrderPlacer, then recording the executed fill in the local ledger.
type LiveExecutor struct {
	Account *LiveAccount
	Placer  OrderPlacer
}

func NewLiveExecutor(account *LiveAccount, placer OrderPlacer) *LiveExecutor {
	return &LiveExecutor{Account: account, Placer: placer}
}

// FeeRate returns the live account's fallback fee rate.
func (le *LiveExecutor) FeeRate() decimal.Decimal { return le.Account.Fee }

// Snapshot exposes the live ledger (AccountSnapshotter).
func (le *LiveExecutor) Snapshot() AccountSnapshot { return le.Account.Snapshot() }

func (le *LiveExecutor) ExecuteFill(order Order) (*Fill, error) {
	if le.Placer == nil {
		return nil, fmt.Errorf("live executor has no order placer configured")
	}

	avgPrice, executedQty, fee, err := le.Placer.PlaceMarketOrder(order.Symbol, string(order.Signal), order.Qty, order.ClientOrderID)
	if err != nil {
		return nil, err
	}
	if fee.IsZero() {
		fee = avgPrice.Mul(executedQty).Mul(le.Account.Fee)
	}

	fill := &Fill{
		BotID:  order.BotID,
		Symbol: order.Symbol,
		Signal: order.Signal,
		Price:  avgPrice,
		Qty:    executedQty,
		Fee:    fee,
		Time:   time.Now(),
	}

	if err := le.Account.ApplyFill(fill); err != nil {
		return nil, err
	}

	return fill, nil
}

func (a *LiveAccount) ApplyFill(f *Fill) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, ok := a.Positions[f.Symbol]; !ok {
		a.Positions[f.Symbol] = decimal.NewFromInt(0)
	}

	switch f.Signal {
	case "BUY":
		totalCost := f.Price.Mul(f.Qty).Add(f.Fee)
		if a.balance.LessThan(totalCost) {
			return fmt.Errorf("not enough cash")
		}
		a.balance = a.balance.Sub(totalCost)
		a.Positions[f.Symbol] = a.Positions[f.Symbol].Add(f.Qty)
		a.costBasis[f.Symbol] = a.costBasis[f.Symbol].Add(f.Price.Mul(f.Qty))

	case "SELL":
		if a.Positions[f.Symbol].LessThan(f.Qty) {
			return fmt.Errorf("not enough asset")
		}
		qtyBefore := a.Positions[f.Symbol]
		avgEntry := decimal.Zero
		if qtyBefore.GreaterThan(decimal.Zero) {
			avgEntry = a.costBasis[f.Symbol].Div(qtyBefore)
		}
		f.Pnl = f.Price.Sub(avgEntry).Mul(f.Qty)
		a.Positions[f.Symbol] = qtyBefore.Sub(f.Qty)
		a.costBasis[f.Symbol] = a.costBasis[f.Symbol].Sub(avgEntry.Mul(f.Qty))
		if a.Positions[f.Symbol].LessThanOrEqual(decimal.Zero) {
			a.Positions[f.Symbol] = decimal.Zero
			a.costBasis[f.Symbol] = decimal.Zero
		}
		a.balance = a.balance.Add(f.Price.Mul(f.Qty).Sub(f.Fee))

	default:
		return fmt.Errorf("invalid side")
	}

	return nil
}

func (a *LiveAccount) Snapshot() AccountSnapshot {
	a.mu.RLock()
	defer a.mu.RUnlock()

	positionsCopy := make(map[string]decimal.Decimal, len(a.Positions))
	maps.Copy(positionsCopy, a.Positions)
	costCopy := make(map[string]decimal.Decimal, len(a.costBasis))
	maps.Copy(costCopy, a.costBasis)

	return AccountSnapshot{Balance: a.balance, Positions: positionsCopy, CostBasis: costCopy}
}
