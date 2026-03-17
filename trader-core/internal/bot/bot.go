package bot

import (
	"context"
	"errors"
	"time"

	"trader-core/internal/db/models"
	"trader-core/internal/dto"
	"trader-core/internal/engine"
	"trader-core/internal/strategies"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type BotStatus string

// Bot statuses
const (
	BotCreated BotStatus = "created"
	BotRunning BotStatus = "running"
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

	Logger *BotLogger

	Status   BotStatus                  `json:"status"`
	Started  time.Time                  `json:"started"`
	Strategy *strategies.SimpleStrategy `json:"strategy"`

	Engine     engine.ExecutionEngine
	MaxCandles int

	CandleCh chan models.Candle
	Candles  []models.Candle

	TradeBroadcaster  *Broadcaster[dto.TradeDTO]
	CandleBroadcaster *Broadcaster[models.Candle]
	TickBroadcaster   *Broadcaster[models.Candle]

	ctx    context.Context
	cancel context.CancelFunc
}

func (b *Bot) Start() error {
	if b.Lookback <= 0 {
		return errors.New("lookback must be > 0")
	}

	for len(b.CandleCh) > 0 {
		candle := <-b.CandleCh
		b.Candles = append(b.Candles, candle)
		if len(b.Candles) > b.MaxCandles {
			b.Candles = b.Candles[len(b.Candles)-b.MaxCandles:]
		}
	}

	b.Started = time.Now()
	b.Status = BotRunning
	b.Logger.Info("started")
	return nil
}

// Recreate the asset symbol
func (b *Bot) Symbol() string {
	return b.Base + b.Quote
}
