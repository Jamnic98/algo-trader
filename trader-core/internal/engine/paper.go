package engine

import (
	"fmt"
	"maps"
	"sync"
	"time"
)

type PaperAccount struct {
	mu        sync.Mutex
	balance   float64
	Positions map[string]float64
	Fee       float64
}

func NewPaperAccount(startBalance, fee float64) *PaperAccount {
	return &PaperAccount{
		balance:   startBalance,
		Fee:       fee,
		Positions: make(map[string]float64),
	}
}

type PaperExecution struct {
	Account *PaperAccount
}

func NewPaperExecution(account *PaperAccount) *PaperExecution {
	return &PaperExecution{Account: account}
}

func (pe *PaperExecution) ExecuteTrade(order Order) (*Fill, error) {
	cost := order.Price * order.Qty
	fee := cost * pe.Account.Fee

	fill := &Fill{
		BotID:  order.BotID,
		Symbol: order.Symbol,
		Side:   order.Side,
		Price:  order.Price,
		Qty:    order.Qty,
		Fee:    fee,
		Time:   time.Now(),
	}

	if err := pe.Account.ApplyFill(fill); err != nil {
		return nil, err
	}

	return fill, nil
}

func (a *PaperAccount) ApplyFill(f *Fill) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	switch f.Side {
	case "BUY":
		if a.balance < f.Price*f.Qty+f.Fee {
			return fmt.Errorf("not enough cash")
		}
		a.balance -= f.Price*f.Qty + f.Fee
		a.Positions[f.Symbol] += f.Qty

	case "SELL":
		if a.Positions[f.Symbol] < f.Qty {
			return fmt.Errorf("not enough asset")
		}
		a.Positions[f.Symbol] -= f.Qty
		a.balance += f.Price*f.Qty - f.Fee

	default:
		return fmt.Errorf("invalid side")
	}

	return nil
}

func (a *PaperAccount) Snapshot() (balance float64, positions map[string]float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	positions = make(map[string]float64)
	maps.Copy(positions, a.Positions)

	return a.balance, positions
}
