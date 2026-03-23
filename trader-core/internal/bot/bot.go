package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"trader-core/internal/db/models"
	"trader-core/internal/dto"
	"trader-core/internal/engine"
	"trader-core/internal/strategies"

	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
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
	BotRunning BotStatus = "trading"
)

type BotMode string

const (
	BotModePaper BotMode = "paper"
	BotModeLive  BotMode = "live"
)

type CreateBotStrategy struct {
	Name        string          `json:"name"`
	DisplayName string          `json:"display_name"`
	Params      json.RawMessage `json:"params"`    // pass through as-is
	MakerFee    *float64        `json:"maker_fee"` // nil = use default
	TakerFee    *float64        `json:"taker_fee"` // nil = use default
}

func (s CreateBotStrategy) ToStrategyConfig(botID string) models.StrategyConfig {
	params := datatypes.JSON([]byte("{}"))
	if s.Params != nil {
		params = datatypes.JSON(s.Params)
	}

	maker := decimal.NewFromFloat(0.001) // default Binance spot
	taker := decimal.NewFromFloat(0.001)
	if s.MakerFee != nil {
		maker = decimal.NewFromFloat(*s.MakerFee)
	}
	if s.TakerFee != nil {
		taker = decimal.NewFromFloat(*s.TakerFee)
	}

	return models.StrategyConfig{
		BotID:       botID,
		Name:        s.Name,
		DisplayName: s.DisplayName,
		Params:      params,
		MakerFee:    maker,
		TakerFee:    taker,
	}
}

type CreateBotData struct {
	Mode       BotMode           `json:"mode"`
	Strategy   CreateBotStrategy `json:"strategy"`
	Exchange   string            `json:"exchange"`
	AssetType  string            `json:"assetType"`
	Base       string            `json:"base"`
	Quote      string            `json:"quote"`
	Interval   string            `json:"interval"`
	MaxCandles int               `json:"maxCandles"`
	Quantity   string            `json:"quantity"`
}

// BotConfig — owns all trading params + which strategy to use
type BotConfig struct {
	ID             string                `gorm:"primaryKey"    json:"id"`
	Mode           BotMode               `json:"mode"`
	StrategyConfig models.StrategyConfig `json:"strategy"     gorm:"foreignKey:BotID"`
	Exchange       string                `json:"exchange"`
	AssetType      string                `json:"assetType"`
	Base           string                `json:"base"`
	Quote          string                `json:"quote"`
	Interval       string                `json:"interval"`
	MaxCandles     int                   `gorm:"column:max_candles" json:"maxCandles"`
	Lookback       string                `json:"lookback"`
	Quantity       string                `json:"quantity"`
	DeletedAt      gorm.DeletedAt        `gorm:"index"         json:"-"`
}

func BuildBotConfig(id string, data CreateBotData) (BotConfig, error) {
	interval, err := engine.ParseInterval(data.Interval)
	if err != nil {
		return BotConfig{}, fmt.Errorf("invalid interval %q: %w", data.Interval, err)
	}
	lookback := interval.Lookback(data.MaxCandles).String()

	return BotConfig{
		ID:             id,
		Mode:           data.Mode,
		Exchange:       data.Exchange,
		AssetType:      data.AssetType,
		Base:           data.Base,
		Quote:          data.Quote,
		Interval:       data.Interval,
		MaxCandles:     data.MaxCandles,
		Quantity:       data.Quantity,
		Lookback:       lookback,
		StrategyConfig: data.Strategy.ToStrategyConfig(id),
	}, nil
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
