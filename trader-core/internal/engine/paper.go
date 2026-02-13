package engine

import (
	"fmt"
	"log"
	"time"

	"trader-core/internal/db/models"
)

type PaperExecution struct {
	Portfolio map[string]float64 // symbol -> quantity
	Cash      float64            // available cash in USD, USDT, etc.
	Trades    []models.Trade     // log of executed trades
	Fee       float64            // trading fee rate
}

func NewPaperExecution(initialCash float64, fee float64) *PaperExecution {
	return &PaperExecution{
		Portfolio: make(map[string]float64),
		Cash:      initialCash,
		Fee:       fee,
	}
}

// Execute a simulated trade
func (pe *PaperExecution) ExecuteTrade(symbol string, side string, price float64, qty float64) error {
	cost := price * qty
	feeAmount := cost * pe.Fee

	switch side {
	case "BUY":
		totalCost := cost + feeAmount
		if pe.Cash < totalCost {
			return fmt.Errorf("not enough cash to buy")
		}
		pe.Cash -= totalCost
		pe.Portfolio[symbol] += qty
	case "SELL":
		if pe.Portfolio[symbol] < qty {
			return fmt.Errorf("not enough asset to sell")
		}
		pe.Portfolio[symbol] -= qty
		pe.Cash += cost - feeAmount
	default:
		return fmt.Errorf("invalid side")
	}

	trade := models.Trade{
		Symbol:    symbol,
		Side:      side,
		Price:     price,
		Quantity:  qty,
		Fee:       feeAmount,
		Timestamp: time.Now(),
	}
	pe.Trades = append(pe.Trades, trade)

	log.Println(trade)
	return nil
}
