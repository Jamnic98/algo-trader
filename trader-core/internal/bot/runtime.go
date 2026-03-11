package bot

import (
	"context"
	"fmt"
	"log"
	"trader-core/internal/engine"
	"trader-core/internal/monitoring"
)

type Runtime struct {
	Account       engine.Account
	BotFactory    *BotFactory
	Dispatcher    *Dispatcher
	MarketManager *MarketDataManager
	Messenger     *monitoring.Messenger
	Errors        chan string
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
	log.Printf("fetched %d candles for %s %s (maxCandles=%d)", len(candles), b.Symbol(), b.Interval, b.MaxCandles)
	b.Candles = candles

	rt.Dispatcher.Subscribe(b.Symbol(), b.Interval, b)
	rt.MarketManager.Subscribe(b.Symbol(), b.Interval)

	// goroutine starts here — drains candles, won't trade until Running
	b.ctx, b.cancel = context.WithCancel(context.Background())
	go RunBotStrategy(b.ctx, b)

	b.Status = BotAttached
	return nil
}

func (rt *Runtime) DetachBot(b *Bot) error {
	if b.Status != BotAttached {
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

	b.Status = BotCreated
	return nil
}
