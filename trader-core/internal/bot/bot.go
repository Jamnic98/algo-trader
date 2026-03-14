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
	"gorm.io/gorm"
)

type BotStatus string

const (
	BotCreated  BotStatus = "created"
	BotAttached BotStatus = "attached"
	BotRunning  BotStatus = "running"
)

type BotMode string

const (
	BotModePaper BotMode = "paper"
	BotModeLive  BotMode = "live"
)

type BotConfig struct {
	ID        string          `gorm:"primaryKey" json:"id"`
	Mode      BotMode         `json:"mode"`
	Base      string          `json:"base"`
	Quote     string          `json:"quote"`
	Interval  engine.Interval `json:"interval"`
	Lookback  time.Duration   `json:"lookback"`
	Quantity  decimal.Decimal `json:"quantity"`
	DeletedAt gorm.DeletedAt  `gorm:"index" json:"-"`
}

type Bot struct {
	BotConfig

	Status   BotStatus                  `json:"status"`
	Started  time.Time                  `json:"started"`
	Strategy *strategies.SimpleStrategy `json:"strategy"`

	Engine     engine.ExecutionEngine
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

	// drain candles accumulated while attached and sync
	for len(b.CandleCh) > 0 {
		candle := <-b.CandleCh
		b.Candles = append(b.Candles, candle)
		if len(b.Candles) > b.MaxCandles {
			b.Candles = b.Candles[len(b.Candles)-b.MaxCandles:]
		}
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

func (b *Bot) Symbol() string {
	return b.Base + b.Quote
}
