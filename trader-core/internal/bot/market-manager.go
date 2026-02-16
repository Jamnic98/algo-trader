package bot

import (
	"context"
	"log"
	"strings"
	"sync"

	"trader-core/internal/binance"
	"trader-core/internal/engine"
)

type MarketDataManager struct {
	client     *binance.Client
	dispatcher *Dispatcher

	mu        sync.Mutex
	refCount  map[string]int
	lastClose map[string]int64 // key -> last closed candle timestamp
}

func NewMarketDataManager(client *binance.Client, dispatcher *Dispatcher) *MarketDataManager {
	log.Println("Creating markektManager")
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
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-m.client.Messages():
			candle, key, ok := engine.ParseKline(msg)
			if !ok { // skip if not closed
				continue
			}
			m.dispatcher.Dispatch(key, candle)
		}
	}

}
