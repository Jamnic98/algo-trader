package bot

import (
	"context"
	"fmt"
	"strings"
	"time"
	"trader-core/internal/db/models"
	"trader-core/internal/engine"
	"trader-core/internal/exchange"
	"trader-core/internal/monitoring"
	"trader-core/internal/strategies"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Runtime struct {
	Account       engine.Account // default paper account (back-compat)
	Accounts      *engine.AccountManager
	BotFactory    *BotFactory
	DB            *gorm.DB
	Dispatcher    *Dispatcher
	MarketManager *MarketDataManager
	Messenger     *monitoring.Messenger
	Errors        chan string
	Brokers       map[string]exchange.Broker // venue name -> broker, for market data + live orders
}

// BrokerFor returns the broker for a venue name ("" = default "binance").
func (r *Runtime) BrokerFor(exchangeName string) (exchange.Broker, error) {
	if exchangeName == "" {
		exchangeName = "binance"
	}
	br, ok := r.Brokers[strings.ToLower(exchangeName)]
	if !ok {
		return nil, fmt.Errorf("no broker for exchange %q (registered: %v)", exchangeName, exchange.Names())
	}
	return br, nil
}

func (r *Runtime) newBot(cfg BotConfig, strategy strategies.Strategy) (*Bot, error) {
	switch cfg.Mode {
	case BotModeLive:
		// live bots place real orders on their own exchange; the per-exchange
		// live ledger is synced from that venue's balances
		br, err := r.BrokerFor(cfg.Exchange)
		if err != nil {
			return nil, err
		}
		quote := strings.ToUpper(cfg.Quote)
		if quote == "" {
			quote = exchange.QuoteAssetFor(cfg.Exchange)
		}
		var liveAcc *engine.LiveAccount
		if r.Accounts != nil {
			bals, berr := br.Balances()
			if berr != nil {
				return nil, fmt.Errorf("fetching %s balances: %w", cfg.Exchange, berr)
			}
			r.Accounts.SyncLive(cfg.Exchange, bals, quote)
			liveAcc = r.Accounts.Live(cfg.Exchange)
		}
		return r.BotFactory.NewLiveBotFor(cfg, strategy, br, liveAcc)
	default:
		// each paper bot trades on its own account (cfg.AccountName, default "default")
		acc, err := r.Accounts.Paper(cfg.AccountName)
		if err != nil {
			return nil, err
		}
		return r.BotFactory.NewBot(cfg, strategy, engine.NewPaperExecutor(acc))
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

	if data.Exchange != "" && !exchange.Known(data.Exchange) {
		return nil, fmt.Errorf("unknown exchange %q (registered: %v)", data.Exchange, exchange.Names())
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
	return cfgs, r.DB.Find(&cfgs).Error
}

func (rt *Runtime) AttachBot(b *Bot) error {
	if b.Status != BotCreated {
		return fmt.Errorf("cannot attach bot from %s", b.Status)
	}

	if rt.MarketManager == nil || !rt.MarketManager.IsRunning() {
		return fmt.Errorf("cannot attach bot: MarketManager not running")
	}

	if !rt.MarketManager.ExchangeReady(b.Exchange) {
		return fmt.Errorf("cannot attach bot: exchange %q market data not connected", b.Exchange)
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

	// risk management: parse limits, start the per-run tracker, wire the
	// kill-switch to detach + notify
	limits, err := b.RiskLimits()
	if err != nil {
		return fmt.Errorf("invalid risk config: %w", err)
	}
	b.Risk = engine.NewRiskState(limits)

	// fetch the venue's trading rules once per attach so orders can be
	// validated client-side (lot size / tick / min notional)
	if br, berr := rt.BrokerFor(b.Exchange); berr == nil {
		if f, ferr := br.Filters(b.Symbol()); ferr == nil {
			b.Filters = &f
			b.Logger.Info("filters %s: minQty=%s stepQty=%s tick=%s minNotional=%s",
				b.Symbol(), f.MinQty.String(), f.StepQty.String(), f.TickSize.String(), f.MinNotional.String())
		} else {
			b.Logger.Warn("could not fetch filters for %s: %v (orders unvalidated)", b.Symbol(), ferr)
		}
	}

	b.OnKillSwitch = func(reason string) {
		if rt.Messenger != nil {
			rt.Messenger.NotifyError(fmt.Sprintf("bot %s (%s %s) kill switch: %s", b.ID, b.Exchange, b.Symbol(), reason))
		}
		if err := rt.DetachBot(b); err != nil {
			b.Logger.Warn("detach after kill switch failed: %v", err)
		}
	}
	b.OnError = func(msg string) {
		if rt.Messenger != nil {
			rt.Messenger.NotifyError(fmt.Sprintf("bot %s (%s %s): %s", b.ID, b.Exchange, b.Symbol(), msg))
		}
	}

	// Subscribe BEFORE fetching history so no closed candle is missed in the gap;
	// afterwards drop stream candles already covered by the fetched history.
	rt.Dispatcher.Subscribe(b.Symbol(), interval, b)
	b.Logger.Info("subscribed to %s_%s", b.Symbol(), interval.String())
	if err := rt.MarketManager.Subscribe(b.Exchange, b.Symbol(), interval); err != nil {
		rt.Dispatcher.Unsubscribe(b.Symbol(), interval, b)
		return fmt.Errorf("failed to subscribe market stream: %w", err)
	}

	candles, err := rt.MarketManager.FetchCandles(b.Exchange, b.Symbol(), interval, b.MaxCandles+1)
	if err != nil {
		rt.Dispatcher.Unsubscribe(b.Symbol(), interval, b)
		rt.MarketManager.Unsubscribe(b.Exchange, b.Symbol(), interval)
		return fmt.Errorf("failed to fetch historical candles: %w", err)
	}
	b.Logger.Info("fetched %d candles", len(candles))
	b.Candles = candles
	if n := len(candles); n > 0 {
		drainCandles(b.CandleCh, candles[n-1].CloseTime)
	}

	// Run the bot
	b.ctx, b.cancel = context.WithCancel(context.Background())
	go RunBotStrategy(b.ctx, b)

	b.Started = time.Now()
	b.Status = BotRunning
	b.Logger.Info("started")
	return nil
}

// drainCandles removes buffered candles with CloseTime <= lastClose (already
// covered by fetched history) and requeues any newer ones, preserving order.
func drainCandles(ch chan models.Candle, lastClose int64) {
	var keep []models.Candle
	for {
		select {
		case c := <-ch:
			if c.CloseTime > lastClose {
				keep = append(keep, c)
			}
		default:
			for _, c := range keep {
				ch <- c
			}
			return
		}
	}
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
	rt.MarketManager.Unsubscribe(b.Exchange, b.Symbol(), interval)

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
	if err := r.DB.First(&strategyModel, "slug = ?", cfg.StrategySlug).Error; err != nil {
		return nil, fmt.Errorf("strategy %s not found: %w", cfg.StrategySlug, err)
	}

	activeStrategy, err := ResolveStrategy(r.DB, strategyModel)
	if err != nil {
		return nil, err
	}
	return r.newBot(cfg, activeStrategy)
}
