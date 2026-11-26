package engine

import (
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type PaperAccount struct {
	balance    decimal.Decimal            // total available funds
	Fee        decimal.Decimal            // fee rate, e.g. 0.001
	Positions  map[string]decimal.Decimal // symbol -> qty
	costBasis  map[string]decimal.Decimal // symbol -> quote spent (avg entry base)
	mu         sync.RWMutex               // thread-safety

	// OnChange is called after every ledger mutation (fills, reset, top-up)
	// so an AccountManager can persist the balance.
	OnChange func()
}

// SetBalance replaces the available balance.
func (a *PaperAccount) SetBalance(amount decimal.Decimal) {
	a.mu.Lock()
	a.balance = amount
	a.mu.Unlock()
	a.notifyChange()
}

// AddBalance tops up the available balance.
func (a *PaperAccount) AddBalance(amount decimal.Decimal) {
	a.mu.Lock()
	a.balance = a.balance.Add(amount)
	a.mu.Unlock()
	a.notifyChange()
}

// Reset clears positions and restores the balance/fee.
func (a *PaperAccount) Reset(balance, fee decimal.Decimal) {
	a.mu.Lock()
	a.balance = balance
	a.Fee = fee
	a.Positions = make(map[string]decimal.Decimal)
	a.costBasis = make(map[string]decimal.Decimal)
	a.mu.Unlock()
	a.notifyChange()
}

func (a *PaperAccount) notifyChange() {
	if a.OnChange != nil {
		a.OnChange()
	}
}

func NewPaperAccount(startBalance, fee string) *PaperAccount {
	return &PaperAccount{
		balance:   decimal.RequireFromString(startBalance),
		Fee:       decimal.RequireFromString(fee),
		Positions: make(map[string]decimal.Decimal),
		costBasis: make(map[string]decimal.Decimal),
	}
}

// RestorePositions loads persisted positions/cost basis after a restart
// (no OnChange — we're rehydrating, not mutating).
func (a *PaperAccount) RestorePositions(positions, costBasis map[string]decimal.Decimal) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if positions != nil {
		a.Positions = positions
	}
	if costBasis != nil {
		a.costBasis = costBasis
	}
}

type PaperExecutor struct {
	Account *PaperAccount
}

func NewPaperExecutor(account *PaperAccount) *PaperExecutor {
	return &PaperExecutor{Account: account}
}

// FeeRate returns the paper account's fee rate.
func (pe *PaperExecutor) FeeRate() decimal.Decimal { return pe.Account.Fee }

// Snapshot exposes the paper ledger (AccountSnapshotter).
func (pe *PaperExecutor) Snapshot() AccountSnapshot { return pe.Account.Snapshot() }

// Fill execution with paper trading engine
func (pe *PaperExecutor) ExecuteFill(order Order) (*Fill, error) {
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

func (a *PaperAccount) ApplyFill(f *Fill) error {
	a.mu.Lock()

	// initialize position if missing
	if _, ok := a.Positions[f.Symbol]; !ok {
		a.Positions[f.Symbol] = decimal.NewFromInt(0)
	}

	var err error
	switch f.Signal {
	case "BUY":
		totalCost := f.Price.Mul(f.Qty).Add(f.Fee)
		if a.balance.LessThan(totalCost) {
			err = fmt.Errorf("not enough cash")
		} else {
			a.balance = a.balance.Sub(totalCost)
			a.Positions[f.Symbol] = a.Positions[f.Symbol].Add(f.Qty)
			a.costBasis[f.Symbol] = a.costBasis[f.Symbol].Add(f.Price.Mul(f.Qty))
		}

	case "SELL":
		if a.Positions[f.Symbol].LessThan(f.Qty) {
			err = fmt.Errorf("not enough asset")
		} else {
			qtyBefore := a.Positions[f.Symbol]
			avgEntry := decimal.Zero
			if qtyBefore.GreaterThan(decimal.Zero) {
				avgEntry = a.costBasis[f.Symbol].Div(qtyBefore)
			}
			f.Pnl = f.Price.Sub(avgEntry).Mul(f.Qty) // realized pnl, ex-fee
			a.Positions[f.Symbol] = qtyBefore.Sub(f.Qty)
			a.costBasis[f.Symbol] = a.costBasis[f.Symbol].Sub(avgEntry.Mul(f.Qty))
			if a.Positions[f.Symbol].LessThanOrEqual(decimal.Zero) {
				a.Positions[f.Symbol] = decimal.Zero
				a.costBasis[f.Symbol] = decimal.Zero
			}
			a.balance = a.balance.Add(f.Price.Mul(f.Qty).Sub(f.Fee))
		}

	default:
		err = fmt.Errorf("invalid side")
	}

	// unlock BEFORE notifyChange: OnChange (persist) takes RLock via Snapshot,
	// which would deadlock if called with the write lock held
	a.mu.Unlock()

	if err != nil {
		return err
	}
	a.notifyChange()
	return nil
}

func (a *PaperAccount) Snapshot() AccountSnapshot {
	a.mu.RLock()
	defer a.mu.RUnlock()

	positionsCopy := make(map[string]decimal.Decimal, len(a.Positions))
	maps.Copy(positionsCopy, a.Positions)
	costCopy := make(map[string]decimal.Decimal, len(a.costBasis))
	maps.Copy(costCopy, a.costBasis)

	return AccountSnapshot{Balance: a.balance, Positions: positionsCopy, CostBasis: costCopy}
}
