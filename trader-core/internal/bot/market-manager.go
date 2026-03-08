package bot

import (
	"context"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"trader-core/internal/binance"
	"trader-core/internal/db/models"
	"trader-core/internal/engine"

	"github.com/shopspring/decimal"
)

type MarketDataManager struct {
	client     *binance.Client
	dispatcher *Dispatcher

	mu        sync.Mutex
	refCount  map[string]int
	lastClose map[string]int64 // key -> last closed candle timestamp
	running   atomic.Bool
}

func NewMarketDataManager(client *binance.Client, dispatcher *Dispatcher) *MarketDataManager {
	log.Println("Creating marketManager")
	return &MarketDataManager{
		client:     client,
		dispatcher: dispatcher,
		refCount:   make(map[string]int),
		lastClose:  make(map[string]int64),
	}
}

func (m *MarketDataManager) Subscribe(symbol string, interval engine.Interval) {
	key := symbol + "_" + interval.String()
	stream := strings.ToLower(symbol) + "@kline_" + interval.String()

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.refCount[key] == 0 {
		m.client.Subscribe(stream)
	}
	m.refCount[key]++
}

func (m *MarketDataManager) Unsubscribe(symbol string, interval engine.Interval) {
	key := symbol + "_" + interval.String()
	stream := strings.ToLower(symbol) + "@kline_" + interval.String()

	m.mu.Lock()
	defer m.mu.Unlock()

	m.refCount[key]--
	if m.refCount[key] <= 0 {
		delete(m.refCount, key)
		m.client.Unsubscribe(stream)
		delete(m.lastClose, key)
	}
}

func (m *MarketDataManager) Run(ctx context.Context) {
	m.running.Store(true)
	defer func() {
		m.running.Store(false)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-m.client.Messages():
			candle, key, ok := engine.ParseKline(msg)
			if !ok {
				continue
			}

			m.mu.Lock()
			m.lastClose[key] = candle.CloseTime
			m.mu.Unlock()

			m.dispatcher.Dispatch(key, candle)
		}
	}
}

func (m *MarketDataManager) IsRunning() bool {
	return m.running.Load()
}

func (m *MarketDataManager) FetchCandles(symbol string, interval engine.Interval, limit int) ([]models.Candle, error) {
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
