package bot

import (
	"context"
	"fmt"
	"time"
	"trader-core/internal/engine"
	"trader-core/internal/monitoring"

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

func (r *Runtime) newBot(cfg BotConfig) (*Bot, error) {
	switch cfg.Mode {
	// TODO: live implementation
	// case BotModeLive:
	// 	return r.BotFactory.NewLiveBot(cfg)
	default:
		return r.BotFactory.NewPaperBot(cfg)
	}
}

func (r *Runtime) CreateBot(cfg BotConfig) (*Bot, error) {
	b, err := r.newBot(cfg)
	if err != nil {
		return nil, err
	}
	b.Logger.Info("created")
	return b, r.DB.Save(&b.BotConfig).Error
}

func (r *Runtime) RestoreBot(cfg BotConfig) (*Bot, error) {
	return r.newBot(cfg)
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
	return cfgs, r.DB.Find(&cfgs).Error
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

	intervalDur := b.Interval.Duration()
	b.MaxCandles = max(int(b.Lookback/intervalDur), 1)

	candles, err := rt.MarketManager.FetchCandles(b.Symbol(), b.Interval, b.MaxCandles+1)
	if err != nil {
		return fmt.Errorf("failed to fetch historical candles: %w", err)
	}
	b.Logger.Info("fetched %d candles", len(candles))
	b.Candles = candles

	rt.Dispatcher.Subscribe(b.Symbol(), b.Interval, b)
	b.Logger.Info("subscribed to %s_%s", b.Symbol(), b.Interval.String())
	rt.MarketManager.Subscribe(b.Symbol(), b.Interval)

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

	rt.Dispatcher.Unsubscribe(b.Symbol(), b.Interval, b)
	rt.MarketManager.Unsubscribe(b.Symbol(), b.Interval)

	if b.cancel != nil {
		b.cancel()
	}
	b.cancel = nil
	b.ctx = nil
	b.Candles = nil

	// drain and recreate
	for len(b.CandleCh) > 0 {
		<-b.CandleCh
	}

	b.Status = BotCreated
	b.Started = time.Time{}
	b.Logger.Info("detached")
	return nil
}
