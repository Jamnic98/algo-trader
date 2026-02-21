package engine

import (
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type PaperAccount struct {
	balance   decimal.Decimal            // total available funds
	Fee       decimal.Decimal            // fee rate, e.g. 0.001
	Positions map[string]decimal.Decimal // symbol -> qty
	mu        sync.RWMutex               // thread-safety
}

func NewPaperAccount(startBalance, fee string) *PaperAccount {
	return &PaperAccount{
		balance:   decimal.RequireFromString(startBalance),
		Fee:       decimal.RequireFromString(fee),
		Positions: make(map[string]decimal.Decimal),
	}
}

type PaperExecution struct {
	Account *PaperAccount
}

func NewPaperExecution(account *PaperAccount) *PaperExecution {
	return &PaperExecution{Account: account}
}

func (pe *PaperExecution) ExecuteTrade(order Order) (*Fill, error) {
	price := order.Price
	qty := order.Qty
	feeRate := pe.Account.Fee

	notional := price.Mul(qty)

	// Fee = notional * feeRate
	fee := notional.Mul(feeRate)

	// Create fill
	fill := &Fill{
		BotID:    order.BotID,
		Symbol:   order.Symbol,
		Side:     order.Side,
		Price:    price,
		Qty:      qty,
		Notional: notional,
		Fee:      fee,
		Time:     time.Now(),
	}

	// Apply to account (thread-safe, decimal-safe)
	if err := pe.Account.ApplyFill(fill); err != nil {
		return nil, err
	}

	return fill, nil
}

func (a *PaperAccount) ApplyFill(f *Fill) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// initialize position if missing
	if _, ok := a.Positions[f.Symbol]; !ok {
		a.Positions[f.Symbol] = decimal.NewFromInt(0)
	}

	switch f.Side {
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

func (a *PaperAccount) Snapshot() AccountSnapshot {
	a.mu.RLock()
	defer a.mu.RUnlock()

	positionsCopy := make(map[string]decimal.Decimal, len(a.Positions))
	maps.Copy(positionsCopy, a.Positions)

	return AccountSnapshot{a.balance, positionsCopy}
}
