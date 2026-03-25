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

type AssetType string

const (
	Crypto AssetType = "crypto"
	Stocks AssetType = "stocks"
)

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

var (
	defaultMakerFee = decimal.NewFromFloat(0.001)
	defaultTakerFee = decimal.NewFromFloat(0.001)
)

// CreateBotRequest is what the HTTP request body looks like.
// The user picks an existing strategy by ID and optionally overrides params/fees.
type CreateBotData struct {
	Mode           BotMode          `json:"mode"`
	StrategyID     uint             `json:"strategy_id"`
	ParamOverrides json.RawMessage  `json:"param_overrides"` // optional, overrides strategy defaults
	MakerFee       *decimal.Decimal `json:"maker_fee"`       // nil = use strategy default
	TakerFee       *decimal.Decimal `json:"taker_fee"`       // nil = use strategy default
	Exchange       string           `json:"exchange"`
	AssetType      string           `json:"assetType"`
	Base           string           `json:"base"`
	Quote          string           `json:"quote"`
	Interval       string           `json:"interval"`
	MaxCandles     int              `json:"maxCandles"`
	Quantity       string           `json:"quantity"`
}

// BotConfig is the DB model for a bot.
type BotConfig struct {
	ID             string          `gorm:"primaryKey"         json:"id"`
	Mode           BotMode         `json:"mode"`
	StrategyID     uint            `gorm:"not null"           json:"strategy_id"`
	Strategy       models.Strategy `gorm:"foreignKey:StrategyID" json:"-"` // loaded via Preload
	ParamOverrides datatypes.JSON  `gorm:"type:jsonb"         json:"param_overrides"`
	Exchange       string          `json:"exchange"`
	AssetType      string          `json:"assetType"`
	Base           string          `json:"base"`
	Quote          string          `json:"quote"`
	Interval       string          `json:"interval"`
	MaxCandles     int             `gorm:"column:max_candles" json:"maxCandles"`
	Lookback       string          `json:"lookback"`
	Quantity       string          `json:"quantity"`
	DeletedAt      gorm.DeletedAt  `gorm:"index"              json:"-"`
}

// ResolvedStrategy merges the strategy template defaults with bot-level overrides.
// func (b *BotConfig) ResolvedStrategy() (dto.StrategyConfigDTO, error) {
// 	// start from strategy template defaults
// 	var cfg dto.StrategyConfigDTO
// 	if err := json.Unmarshal(b.Strategy.DefaultConfig, &cfg); err != nil {
// 		return cfg, fmt.Errorf("parsing strategy default config: %w", err)
// 	}

// 	// apply param overrides if present
// 	if len(b.ParamOverrides) > 0 {
// 		var overrides map[string]any
// 		if err := json.Unmarshal(b.ParamOverrides, &overrides); err != nil {
// 			return cfg, fmt.Errorf("parsing param overrides: %w", err)
// 		}
// 		for k, v := range overrides {
// 			cfg.Params[k] = v
// 		}
// 	}

// 	return cfg, nil
// }

func BuildBotConfig(id string, data CreateBotData) (BotConfig, error) {
	interval, err := engine.ParseInterval(data.Interval)
	if err != nil {
		return BotConfig{}, fmt.Errorf("invalid interval %q: %w", data.Interval, err)
	}
	lookback := interval.Lookback(data.MaxCandles).String()

	// normalise param overrides
	var overrides datatypes.JSON
	if len(data.ParamOverrides) > 0 {
		overrides = datatypes.JSON(data.ParamOverrides)
	}

	return BotConfig{
		ID:             id,
		Mode:           data.Mode,
		StrategyID:     data.StrategyID,
		ParamOverrides: overrides,
		Exchange:       data.Exchange,
		AssetType:      data.AssetType,
		Base:           data.Base,
		Quote:          data.Quote,
		Interval:       data.Interval,
		MaxCandles:     data.MaxCandles,
		Quantity:       data.Quantity,
		Lookback:       lookback,
	}, nil
}

// Bot is the runtime struct — BotConfig is the DB side, everything else is in-memory.
type Bot struct {
	BotConfig

	Logger         *BotLogger
	Status         BotStatus           `json:"status"`
	Started        time.Time           `json:"started"`
	ActiveStrategy strategies.Strategy // renamed to avoid clash with BotConfig.Strategy (the DB relation)

	Engine engine.ExecutionEngine

	CandleCh chan models.Candle
	Candles  []models.Candle

	TradeBroadcaster  *Broadcaster[dto.TradeDTO]
	CandleBroadcaster *Broadcaster[models.Candle]
	TickBroadcaster   *Broadcaster[models.Candle]

	ctx    context.Context
	cancel context.CancelFunc
}

func (b *Bot) Symbol() string {
	switch b.Exchange {
	case "binance":
		return strings.ToUpper(b.Base + b.Quote)
	}
	return strings.ToUpper(b.Base + b.Quote)
}
