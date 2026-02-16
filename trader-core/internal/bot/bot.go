package bot

import (
	"context"
	"errors"
	"time"

	"trader-core/internal/db/models"
	"trader-core/internal/engine"
	"trader-core/internal/strategies"
)

type Bot struct {
	ID         string                    `json:"id"`
	Interval   Interval                  `json:"interval"`
	Symbol     string                    `json:"symbol"`
	Status     string                    `json:"status"`
	Started    time.Time                 `json:"started"`
	Strategy   strategies.SimpleStrategy `json:"strategy"`
	Engine     engine.TradeExecutor
	Lookback   time.Duration
	MaxCandles int

	CandleCh chan models.Candle
	Candles  []models.Candle

	Ctx    context.Context
	cancel context.CancelFunc
}

func (b *Bot) Start() error {
	if b.Ctx != nil {
		return errors.New("bot already running")
	}

	if b.Lookback <= 0 {
		return errors.New("lookback must be > 0")
	}

	intervalDur := b.Interval.Duration()
	b.MaxCandles = max(int(b.Lookback/intervalDur), 1)

	b.Ctx, b.cancel = context.WithCancel(context.Background())
	b.Status = "running"
	b.Started = time.Now()

	go RunBotStrategy(b.Ctx, b)
	return nil
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
