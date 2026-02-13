package bot

import (
	"context"
	"time"

	"trader-core/internal/db/models"
	"trader-core/internal/engine"
)

type Bot struct {
	ID       string         `json:"id"`
	Interval string         `json:"interval"`
	Symbol   string         `json:"symbol"`
	Status   string         `json:"status"`
	Started  time.Time      `json:"started"`
	Strategy SimpleStrategy `json:"strategy"`
	Engine   engine.TradeExecutor

	CandleCh    chan models.Candle
	Candles     []models.Candle
	MaxLookback int

	cancel context.CancelFunc
}

func (b *Bot) Stop() {
	if b.cancel != nil {
		b.cancel()
		b.Status = "stopped"
	}
}

func (b *Bot) SetCancel(c context.CancelFunc) {
	b.cancel = c
}

type SimpleStrategy struct {
	hasPosition bool
}

func NewSimpleStrategy() *SimpleStrategy {
	return &SimpleStrategy{}
}

type Side string

const (
	BUY  Side = "BUY"
	SELL Side = "SELL"
	NONE Side = "NONE"
)

func (s SimpleStrategy) OnCandles(candles []models.Candle) Side {
	last := candles[len(candles)-1]
	prev := candles[len(candles)-2]

	if last.Close > prev.Close {
		return BUY
	}

	if last.Close < prev.Close {
		return SELL
	}

	return NONE
}
