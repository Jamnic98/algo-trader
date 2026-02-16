package engine

import (
	"fmt"
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

func (pe *PaperExecution) ExecuteTrade(botID, symbol string, side Side, price, qty float64) (*Fill, error) {
	cost := price * qty
	fee := cost * pe.Account.Fee

	fill := &Fill{
		BotID:  botID,
		Symbol: symbol,
		Side:   side,
		Price:  price,
		Qty:    qty,
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
