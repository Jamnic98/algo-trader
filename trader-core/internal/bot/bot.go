package bot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"trader-core/internal/db/models"
	"trader-core/internal/dto"
	"trader-core/internal/engine"
	"trader-core/internal/strategies"

	"github.com/shopspring/decimal"
)

type BotStatus string

const (
	BotCreated  BotStatus = "created"
	BotAttached BotStatus = "attached"
	BotRunning  BotStatus = "running"
)

type BotConfig struct {
	ID       string          `gorm:"primaryKey" json:"id"`
	Symbol   string          `json:"symbol"`
	Interval engine.Interval `json:"interval"`
	Lookback time.Duration   `json:"lookback"`
	Quantity decimal.Decimal `json:"quantity"`
}

type Bot struct {
	ID       string                     `json:"id"`
	Interval engine.Interval            `json:"interval"`
	Symbol   string                     `json:"symbol"`
	Status   BotStatus                  `json:"status"`
	Started  time.Time                  `json:"started"`
	Strategy *strategies.SimpleStrategy `json:"strategy"`
	Quantity decimal.Decimal            `json:"quantity"`

	Engine     engine.ExecutionEngine
	Lookback   time.Duration
	MaxCandles int

	CandleCh chan models.Candle
	Candles  []models.Candle

	TradeBroadcaster *Broadcaster[dto.TradeDTO]

	ctx    context.Context
	cancel context.CancelFunc
}

func (b *Bot) Start() error {
	if b.Status != BotAttached {
		return fmt.Errorf("cannot start bot from %s", b.Status)
	}

	if b.Lookback <= 0 {
		return errors.New("lookback must be > 0")
	}

	// drain candles accumulated while attached
	for len(b.CandleCh) > 0 {
		<-b.CandleCh
	}

	b.Started = time.Now()
	b.Status = BotRunning
	return nil
}

func (b *Bot) Stop() {
	if b.Status != BotRunning {
		return
	}
	b.Status = BotAttached
	b.Started = time.Time{}
}

// TODO: review usage
func (b *Bot) SetCancel(c context.CancelFunc) {
	b.cancel = c
}
