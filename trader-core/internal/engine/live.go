package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type LiveAccount struct {
	balance   decimal.Decimal            // total available funds
	Fee       decimal.Decimal            // fee rate, e.g. 0.001
	Positions map[string]decimal.Decimal // symbol -> qty
	mu        sync.RWMutex               // thread-safety
}

func NewLiveAccount(startBalance, fee string) *LiveAccount {
	return &LiveAccount{
		balance:   decimal.RequireFromString(startBalance),
		Fee:       decimal.RequireFromString(fee),
		Positions: make(map[string]decimal.Decimal),
	}
}

type LiveExecutor struct {
	Account *LiveAccount
}

func NewLiveExecutor(account *LiveAccount) *LiveExecutor {
	return &LiveExecutor{Account: account}
}

// Fill execution with Live trading engine
func (pe *LiveExecutor) ExecuteFill(order Order) (*Fill, error) {
	price := order.Price
	qty := order.Qty
	feeRate := pe.Account.Fee

	// Calculate notional (Price * Quantity)
	notional := price.Mul(qty)
	// Fee amount = notional * feeRate
	fee := notional.Mul(feeRate)

	// Create fill
	fill := &Fill{
		BotID:  order.BotID,
		Symbol: order.Symbol,
		Signal: order.Signal,
		Price:  price,
		Fee:    fee,
		Qty:    qty,
		Time:   time.Now(),
	}

	// Apply to account
	if err := pe.Account.ApplyFill(fill); err != nil {
		return nil, err
	}

	return fill, nil
}

func (a *LiveAccount) ApplyFill(f *Fill) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// initialize position if missing
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

	case "SELL":
		if a.Positions[f.Symbol].LessThan(f.Qty) {
			return fmt.Errorf("not enough asset")
		}
		a.Positions[f.Symbol] = a.Positions[f.Symbol].Sub(f.Qty)
		a.balance = a.balance.Add(f.Price.Mul(f.Qty).Sub(f.Fee))

	default:
		return fmt.Errorf("invalid side")
	}

	return nil
}

// func (a *LiveAccount) Snapshot() AccountSnapshot {
// 	a.mu.RLock()
// 	defer a.mu.RUnlock()

// 	positionsCopy := make(map[string]decimal.Decimal, len(a.Positions))
// 	maps.Copy(positionsCopy, a.Positions)

// 	return AccountSnapshot{a.balance, positionsCopy}
// }
