package bot

import (
	"context"
	"fmt"
	"time"
	"trader-core/internal/db/models"
	"trader-core/internal/engine"
	"trader-core/internal/monitoring"
	"trader-core/internal/strategies"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Runtime struct {
	Account       engine.Account
	BotFactory    *BotFactory
	DB            *gorm.DB
	Dispatcher    *Dispatcher
	MarketManager *MarketDataManager
	Messenger     *monitoring.Messenger
	Errors        chan string
}

func (r *Runtime) newBot(cfg BotConfig, strategy strategies.Strategy) (*Bot, error) {
	switch cfg.Mode {
	default:
		return r.BotFactory.NewPaperBot(cfg, strategy)
	}
}

func (r *Runtime) CreateBot(data CreateBotData) (*Bot, error) {
	var strategyModel models.Strategy

	if err := r.DB.First(&strategyModel, "slug = ?", data.StrategySlug).Error; err != nil {
		return nil, fmt.Errorf("strategy %q not found: %w", data.StrategySlug, err)
	}

	activeStrategy, err := ResolveStrategy(r.DB, strategyModel)
	if err != nil {
		return nil, fmt.Errorf("resolving strategy: %w", err)
	}

	id := uuid.New().String()
	cfg, err := BuildBotConfig(id, data)
	if err != nil {
		return nil, err
	}

	b, err := r.newBot(cfg, activeStrategy)
	if err != nil {
		return nil, err
	}

	b.StrategyName = strategyModel.DisplayName

	if err := r.DB.Create(&b.BotConfig).Error; err != nil {
		return nil, fmt.Errorf("failed to create bot config: %w", err)
	}

	b.Logger.Info("new bot created")
	return b, nil
}

func (r *Runtime) DeleteBot(b *Bot) error {
	result := r.DB.Delete(&BotConfig{}, "id = ?", b.ID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("bot %s not found in db", b.ID)
	}
	b.Logger.Info("deleted")
	return nil
}

func (r *Runtime) LoadBots() ([]BotConfig, error) {
	var cfgs []BotConfig
	return cfgs, r.DB.Preload("Strategy").Find(&cfgs).Error
}

func (rt *Runtime) AttachBot(b *Bot) error {
	if b.Status != BotCreated {
		return fmt.Errorf("cannot attach bot from %s", b.Status)
	}

	if rt.MarketManager == nil || !rt.MarketManager.IsRunning() {
		return fmt.Errorf("cannot attach bot: MarketManager not running")
	}

	if !rt.MarketManager.client.IsAlive() {
		return fmt.Errorf("cannot attach bot: Binance client not connected")
	}

	// Parse interval + lookback from config strings at attach time
	interval, err := engine.ParseInterval(b.Interval)
	if err != nil {
		return fmt.Errorf("invalid interval %q: %w", b.Interval, err)
	}
	lookback, err := time.ParseDuration(b.Lookback)
	if err != nil {
		return fmt.Errorf("invalid lookback %q: %w", b.Lookback, err)
	}

	// Set interval between candle
	intervalDur := interval.Duration()
	b.MaxCandles = max(int(lookback/intervalDur), 1)

	candles, err := rt.MarketManager.FetchCandles(b.Symbol(), interval, b.MaxCandles+1)
	if err != nil {
		return fmt.Errorf("failed to fetch historical candles: %w", err)
	}
	b.Logger.Info("fetched %d candles", len(candles))
	b.Candles = candles

	// Subscribe to OHLC_Candle stream
	rt.Dispatcher.Subscribe(b.Symbol(), interval, b)
	b.Logger.Info("subscribed to %s_%s", b.Symbol(), interval.String())
	rt.MarketManager.Subscribe(b.Symbol(), interval)

	// Run the bot
	b.ctx, b.cancel = context.WithCancel(context.Background())
	go RunBotStrategy(b.ctx, b)

	b.Started = time.Now()
	b.Status = BotRunning
	b.Logger.Info("started")
	return nil
}

func (rt *Runtime) DetachBot(b *Bot) error {
	if b.Status != BotRunning {
		return fmt.Errorf("cannot detach bot from %s", b.Status)
	}

	if rt.MarketManager == nil || !rt.MarketManager.IsRunning() {
		return fmt.Errorf("cannot detach bot: MarketManager not running")
	}

	interval, err := engine.ParseInterval(b.Interval)
	if err != nil {
		return fmt.Errorf("invalid interval %q: %w", b.Interval, err)
	}

	rt.Dispatcher.Unsubscribe(b.Symbol(), interval, b)
	rt.MarketManager.Unsubscribe(b.Symbol(), interval)

	if b.cancel != nil {
		b.cancel()
	}
	b.cancel = nil
	b.ctx = nil
	b.Candles = nil

	for len(b.CandleCh) > 0 {
		<-b.CandleCh
	}

	b.Status = BotCreated
	b.Started = time.Time{}
	b.Logger.Info("detached")
	return nil
}

func (r *Runtime) RestoreBot(cfg BotConfig) (*Bot, error) {
	var strategyModel models.Strategy
	if err := r.DB.First(&strategyModel, cfg.StrategySlug).Error; err != nil {
		return nil, fmt.Errorf("strategy %s not found: %w", cfg.StrategySlug, err)
	}

	activeStrategy, err := ResolveStrategy(r.DB, strategyModel)
	if err != nil {
		return nil, err
	}
	return r.newBot(cfg, activeStrategy)
}
