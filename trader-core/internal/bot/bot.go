package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	_ "trader-core/internal/alpaca" // register the alpaca venue + symbol mapper
	"trader-core/internal/db/models"
	"trader-core/internal/dto"
	"trader-core/internal/engine"
	"trader-core/internal/exchange"
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
	Account        string           `json:"account"` // paper account name; empty = "default"; "live" for live trading
	StrategySlug   string           `json:"strategy_slug"`
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
	Risk           json.RawMessage  `json:"risk"` // engine.RiskLimits as JSON; empty = defaults
}

// BotConfig is the DB model for a bot.
type BotConfig struct {
	ID             string         `gorm:"primaryKey"         json:"id"`
	Mode           BotMode        `json:"mode"`
	AccountName    string         `gorm:"not null;default:'default'" json:"account"` // paper account name, or "live"
	StrategySlug   string         `gorm:"not null"           json:"strategy_slug"` // loaded via Preload
	ParamOverrides datatypes.JSON `gorm:"type:jsonb"         json:"param_overrides"`
	Exchange       string         `json:"exchange"`
	AssetType      string         `json:"assetType"`
	Base           string         `json:"base"`
	Quote          string         `json:"quote"`
	Interval       string         `json:"interval"`
	MaxCandles     int            `gorm:"column:max_candles" json:"maxCandles"`
	Lookback       string         `json:"lookback"`
	Quantity       string         `json:"quantity"`
	Risk           datatypes.JSON `gorm:"type:jsonb"         json:"risk"` // engine.RiskLimits
	DeletedAt      gorm.DeletedAt `gorm:"index"              json:"-"`
}

// RiskLimits parses the stored risk config, falling back to defaults.
func (c *BotConfig) RiskLimits() (engine.RiskLimits, error) {
	if len(c.Risk) == 0 {
		return engine.DefaultRiskLimits(), nil
	}
	return engine.ParseRiskLimits(c.Risk)
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

	account := data.Account
	if account == "" {
		account = "default"
	}
	exchangeName := data.Exchange
	if exchangeName == "" {
		exchangeName = "binance"
	}

	// risk limits: defaults when empty, validated otherwise
	risk := datatypes.JSON{}
	if len(data.Risk) > 0 {
		if _, err := engine.ParseRiskLimits(data.Risk); err != nil {
			return BotConfig{}, err
		}
		risk = datatypes.JSON(data.Risk)
	} else {
		def, _ := json.Marshal(engine.DefaultRiskLimits())
		risk = datatypes.JSON(def)
	}

	return BotConfig{
		ID:             id,
		Mode:           data.Mode,
		AccountName:    account,
		StrategySlug:   data.StrategySlug,
		ParamOverrides: overrides,
		Exchange:       exchangeName,
		AssetType:      data.AssetType,
		Base:           data.Base,
		Quote:          data.Quote,
		Interval:       data.Interval,
		MaxCandles:     data.MaxCandles,
		Quantity:       data.Quantity,
		Lookback:       lookback,
		Risk:           risk,
	}, nil
}

// Bot is the runtime struct — BotConfig is the DB side, everything else is in-memory.
type Bot struct {
	BotConfig

	Logger         *BotLogger
	Status         BotStatus `json:"status"`
	Started        time.Time `json:"started"`
	StrategyName   string    `json:"strategy_name"`
	ActiveStrategy strategies.Strategy

	Engine engine.ExecutionEngine

	CandleCh chan models.Candle
	Candles  []models.Candle

	FillBroadcaster   *Broadcaster[dto.FIllDTO]
	CandleBroadcaster *Broadcaster[models.Candle]
	TickBroadcaster   *Broadcaster[models.Candle]

	ctx    context.Context
	cancel context.CancelFunc

	// Risk is the per-run risk tracker (created on attach); OnKillSwitch is
	// invoked once when a kill-switch trips (detach + notify).
	Risk         *engine.RiskState
	OnKillSwitch func(reason string)
}

// Symbol returns the venue-native symbol for this bot's pair
// (binance "BTCUSDT", alpaca crypto "BTC/USD", alpaca stocks "AAPL").
func (b *Bot) Symbol() string {
	return exchange.SymbolFor(b.Exchange, b.Base, b.Quote, b.AssetType)
}
