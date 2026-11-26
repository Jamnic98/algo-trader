package bot

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"trader-core/internal/binance"
	"trader-core/internal/db/models"
	"trader-core/internal/engine"
	"trader-core/internal/exchange"

	"github.com/shopspring/decimal"
)

// MarketDataManager routes candle streams to bots. Binance keeps using the
// shared WS client (ticks + auto-reconnect); every other venue streams through
// its exchange.Broker with a per-subscription goroutine.
type MarketDataManager struct {
	client     *binance.Client
	dispatcher *Dispatcher
	brokers    map[string]exchange.Broker

	mu        sync.Mutex
	ctx       context.Context
	refCount  map[string]int
	streams   map[string]context.CancelFunc // broker stream key -> cancel
	lastClose map[string]int64               // key -> last closed candle timestamp
	running   atomic.Bool
}

func NewMarketDataManager(client *binance.Client, dispatcher *Dispatcher, brokers map[string]exchange.Broker) *MarketDataManager {
	log.Println("Creating marketManager")
	if brokers == nil {
		brokers = map[string]exchange.Broker{}
	}
	return &MarketDataManager{
		client:     client,
		dispatcher: dispatcher,
		brokers:    brokers,
		refCount:   make(map[string]int),
		streams:    make(map[string]context.CancelFunc),
		lastClose:  make(map[string]int64),
	}
}

func streamKey(exchangeName, symbol string, interval engine.Interval) string {
	return exchangeName + "|" + symbol + "_" + interval.String()
}

// Subscribe starts (or joins) a candle stream for exchangeName+symbol+interval.
func (m *MarketDataManager) Subscribe(exchangeName, symbol string, interval engine.Interval) error {
	key := streamKey(exchangeName, symbol, interval)

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.refCount[key] == 0 {
		if exchangeName == "binance" {
			if m.client != nil {
				stream := strings.ToLower(symbol) + "@kline_" + interval.String()
				m.client.Subscribe(stream)
			}
		} else {
			br, ok := m.brokers[strings.ToLower(exchangeName)]
			if !ok {
				return fmt.Errorf("no broker registered for exchange %q", exchangeName)
			}
			ctx := m.ctx
			if ctx == nil {
				ctx = context.Background()
			}
			sctx, cancel := context.WithCancel(ctx)
			m.streams[key] = cancel
			go m.stream(sctx, br, exchangeName, symbol, interval, key)
		}
	}
	m.refCount[key]++
	return nil
}

// stream pumps klines from a broker into the dispatcher, reconnecting on error.
// Brokers implementing exchange.TickStreamer also get in-progress bars as ticks.
func (m *MarketDataManager) stream(ctx context.Context, br exchange.Broker, exchangeName, symbol string, interval engine.Interval, key string) {
	dispatchKey := symbol + "_" + interval.String()
	toCandle := func(k exchange.Kline) models.Candle {
		return models.Candle{
			OpenTime:  k.OpenTime,
			CloseTime: k.CloseTime,
			Open:      k.Open,
			High:      k.High,
			Low:       k.Low,
			Close:     k.Close,
			Volume:    k.Volume,
		}
	}
	onKline := func(k exchange.Kline) {
		m.mu.Lock()
		m.lastClose[key] = k.CloseTime
		m.mu.Unlock()
		m.dispatcher.Dispatch(dispatchKey, toCandle(k))
	}

	for {
		// backfill any closed candles missed while the stream was down
		m.backfill(exchangeName, symbol, interval, key)

		var err error
		if ts, ok := br.(exchange.TickStreamer); ok {
			err = ts.StreamKlineEvents(ctx, symbol, interval.String(), onKline, func(k exchange.Kline) {
				m.dispatcher.DispatchTick(dispatchKey, toCandle(k))
			})
		} else {
			err = br.StreamKlines(ctx, symbol, interval.String(), onKline)
		}
		if ctx.Err() != nil {
			return
		}
		log.Printf("market stream %s ended: %v; reconnecting", key, err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

// backfill dispatches closed candles with openTime after the last seen close,
// filling the gap left while a stream was disconnected.
func (m *MarketDataManager) backfill(exchangeName, symbol string, interval engine.Interval, key string) {
	m.mu.Lock()
	last := m.lastClose[key]
	m.mu.Unlock()
	if last <= 0 {
		return
	}

	now := time.Now().UnixMilli()
	var raw []exchange.Kline
	var err error
	if exchangeName == "binance" {
		var rk []binance.RawKline
		rk, err = binance.FetchKlinesRange(symbol, interval.String(), last+1, now)
		for _, k := range rk {
			raw = append(raw, exchange.Kline{
				OpenTime: k.OpenTime, CloseTime: k.CloseTime,
				Open: decimal.RequireFromString(k.Open), High: decimal.RequireFromString(k.High),
				Low: decimal.RequireFromString(k.Low), Close: decimal.RequireFromString(k.Close),
				Volume: decimal.RequireFromString(k.Volume),
			})
		}
	} else {
		m.mu.Lock()
		br, ok := m.brokers[strings.ToLower(exchangeName)]
		m.mu.Unlock()
		if !ok {
			return
		}
		raw, err = br.Klines(symbol, interval.String(), last+1, 0, 1000)
	}
	if err != nil {
		log.Printf("backfill %s failed: %v", key, err)
		return
	}

	dispatchKey := symbol + "_" + interval.String()
	for _, k := range raw {
		if k.CloseTime > now {
			continue // still open
		}
		m.dispatcher.Dispatch(dispatchKey, models.Candle{
			OpenTime: k.OpenTime, CloseTime: k.CloseTime,
			Open: k.Open, High: k.High, Low: k.Low, Close: k.Close, Volume: k.Volume,
		})
		m.mu.Lock()
		m.lastClose[key] = k.CloseTime
		m.mu.Unlock()
	}
}

func (m *MarketDataManager) Unsubscribe(exchangeName, symbol string, interval engine.Interval) {
	key := streamKey(exchangeName, symbol, interval)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.refCount[key]--
	if m.refCount[key] <= 0 {
		delete(m.refCount, key)
		delete(m.lastClose, key)
		if exchangeName == "binance" {
			if m.client != nil {
				stream := strings.ToLower(symbol) + "@kline_" + interval.String()
				m.client.Unsubscribe(stream)
			}
		} else if cancel, ok := m.streams[key]; ok {
			cancel()
			delete(m.streams, key)
		}
	}
}

// ExchangeReady reports whether market data can flow for the exchange:
// the Binance WS client must be alive; other venues just need a registered broker.
func (m *MarketDataManager) ExchangeReady(exchangeName string) bool {
	if exchangeName == "binance" {
		return m.client != nil && m.client.IsAlive()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.brokers[strings.ToLower(exchangeName)]
	return ok
}

func (m *MarketDataManager) Run(ctx context.Context) {
	m.mu.Lock()
	m.ctx = ctx
	m.mu.Unlock()
	m.running.Store(true)
	defer func() {
		m.running.Store(false)
	}()

	// nil client (tests / non-binance-only setups) -> nil channel blocks forever
	var msgs <-chan []byte
	if m.client != nil {
		msgs = m.client.Messages()
		// after a binance WS reconnect + resubscribe, backfill missed candles
		m.client.OnReconnect = func() {
			m.mu.Lock()
			keys := make([]string, 0, len(m.refCount))
			for k := range m.refCount {
				keys = append(keys, k)
			}
			m.mu.Unlock()
			for _, k := range keys {
				parts := strings.SplitN(k, "|", 2)
				if len(parts) != 2 || parts[0] != "binance" {
					continue
				}
				si := strings.LastIndex(parts[1], "_")
				if si < 0 {
					continue
				}
				if interval, err := engine.ParseInterval(parts[1][si+1:]); err == nil {
					m.backfill("binance", parts[1][:si], interval, k)
				}
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-msgs:
			// closed candle
			if candle, key, ok := engine.ParseKline(msg); ok {
				m.mu.Lock()
				m.lastClose[key] = candle.CloseTime
				m.mu.Unlock()
				m.dispatcher.Dispatch(key, candle)
			}

			// open candle tick
			if candle, key, ok := engine.ParseKlineTick(msg); ok {
				m.dispatcher.DispatchTick(key, candle)
			}
		}
	}
}

func (m *MarketDataManager) IsRunning() bool {
	return m.running.Load()
}

// FetchCandles loads historical candles for the bot's exchange, skipping the
// still-open last bar.
func (m *MarketDataManager) FetchCandles(exchangeName, symbol string, interval engine.Interval, limit int) ([]models.Candle, error) {
	if exchangeName != "binance" {
		m.mu.Lock()
		br, ok := m.brokers[strings.ToLower(exchangeName)]
		m.mu.Unlock()
		if !ok {
			return nil, fmt.Errorf("no broker registered for exchange %q", exchangeName)
		}
		raw, err := br.Klines(symbol, interval.String(), 0, 0, limit)
		if err != nil {
			return nil, err
		}
		candles := make([]models.Candle, 0, len(raw))
		for _, k := range raw {
			if k.CloseTime > time.Now().UnixMilli() {
				continue
			}
			candles = append(candles, models.Candle{
				OpenTime:  k.OpenTime,
				CloseTime: k.CloseTime,
				Open:      k.Open,
				High:      k.High,
				Low:       k.Low,
				Close:     k.Close,
				Volume:    k.Volume,
			})
		}
		return candles, nil
	}

	raw, err := binance.FetchKlines(symbol, interval.String(), limit)
	if err != nil {
		return nil, err
	}

	candles := make([]models.Candle, 0, len(raw))
	for _, k := range raw {
		// skip the last candle — it's still open/live
		if k.CloseTime > time.Now().UnixMilli() {
			continue
		}
		candles = append(candles, models.Candle{
			OpenTime:  k.OpenTime,
			CloseTime: k.CloseTime,
			Open:      decimal.RequireFromString(k.Open),
			High:      decimal.RequireFromString(k.High),
			Low:       decimal.RequireFromString(k.Low),
			Close:     decimal.RequireFromString(k.Close),
			Volume:    decimal.RequireFromString(k.Volume),
		})
	}

	return candles, nil
}
