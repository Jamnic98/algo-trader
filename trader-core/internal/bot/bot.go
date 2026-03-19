package bot

import (
	"context"
	"strings"
	"time"

	"trader-core/internal/db/models"
	"trader-core/internal/dto"
	"trader-core/internal/engine"
	"trader-core/internal/strategies"

	"gorm.io/gorm"
)

// Asset types
type AssetType string

const (
	Crypto AssetType = "crypto"
	Stocks AssetType = "stocks"
)

// Bot statuses
type BotStatus string

const (
	BotCreated BotStatus = "created"
	BotRunning BotStatus = "running"
)

type BotMode string

const (
	BotModePaper BotMode = "paper"
	BotModeLive  BotMode = "live"
)

type CreateBotData struct {
	Mode       BotMode               `json:"mode"`
	Strategy   models.StrategyConfig `json:"strategy"`
	Exchange   string                `json:"exchange"`
	AssetType  string                `json:"assetType"`
	Base       string                `json:"base"`
	Quote      string                `json:"quote"`
	Interval   string                `json:"interval"`
	MaxCandles int                   `json:"maxCandles"`
	Quantity   string                `json:"quantity"`
}

// BotConfig — owns all trading params + which strategy to use
type BotConfig struct {
	ID             string                `gorm:"primaryKey" json:"id"`
	Mode           BotMode               `json:"mode"`
	StrategyConfig models.StrategyConfig `json:"strategy" gorm:"embedded;embeddedPrefix:strategy_"`
	Exchange       string                `json:"exchange"`
	AssetType      string                `json:"assetType"`
	Base           string                `json:"base"`
	Quote          string                `json:"quote"`
	Interval       string                `json:"interval"`                             // e.g. "1h"
	MaxCandles     int                   `gorm:"column:max_candles" json:"maxCandles"` // e.g. 200
	Lookback       string                `json:"lookback"`                             // e.g. "200h"
	Quantity       string                `json:"quantity"`                             // e.g. "0.001"
	DeletedAt      gorm.DeletedAt        `gorm:"index" json:"-"`
}

type Bot struct {
	BotConfig

	Logger   *BotLogger
	Status   BotStatus `json:"status"`
	Started  time.Time `json:"started"`
	Strategy strategies.Strategy

	Engine engine.ExecutionEngine

	CandleCh chan models.Candle
	Candles  []models.Candle

	TradeBroadcaster  *Broadcaster[dto.TradeDTO]
	CandleBroadcaster *Broadcaster[models.Candle]
	TickBroadcaster   *Broadcaster[models.Candle]

	ctx    context.Context
	cancel context.CancelFunc
}

// Recreate the asset symbol
func (b *Bot) Symbol() string {
	switch b.Exchange {
	case "binance":
		return strings.ToUpper(b.Base + b.Quote) // "BTCUSDT"
	}

	// fallback
	return strings.ToUpper(b.Base + b.Quote)
}
