package binance

import (
	"context"
	"encoding/json"
	"strings"
	"sync"

	"trader-core/internal/db/models"
	"trader-core/internal/engine"
)

type MarketDataManager struct {
	client     *Client
	dispatcher *Dispatcher

	mu       sync.Mutex
	refCount map[string]int
}

type KlineEvent struct {
	EventType string `json:"e"`
	EventTime int64  `json:"E"`
	Symbol    string `json:"s"`
	K         struct {
		StartTime int64       `json:"t"`
		CloseTime int64       `json:"T"`
		Symbol    string      `json:"s"`
		Interval  string      `json:"i"`
		Open      json.Number `json:"o"`
		Close     json.Number `json:"c"`
		High      json.Number `json:"h"`
		Low       json.Number `json:"l"`
		Volume    json.Number `json:"v"`
		Trades    int64       `json:"n"`
		IsClosed  bool        `json:"x"`
	} `json:"k"`
}

func parseKline(msg []byte) (models.Candle, string, bool) {
	var evt KlineEvent
	if err := json.Unmarshal(msg, &evt); err != nil {
		return models.Candle{}, "", false
	}

	if !evt.K.IsClosed {
		return models.Candle{}, "", false
	}

	open, err1 := evt.K.Open.Float64()
	high, err2 := evt.K.High.Float64()
	low, err3 := evt.K.Low.Float64()
	closePrice, err4 := evt.K.Close.Float64()
	volume, err5 := evt.K.Volume.Float64()

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
		return models.Candle{}, "", false
	}

	candle := models.Candle{
		Open:   open,
		High:   high,
		Low:    low,
		Close:  closePrice,
		Volume: volume,
	}

	key := evt.K.Symbol + "_" + evt.K.Interval
	return candle, key, true
}

func NewMarketDataManager(
	client *Client,
	dispatcher *Dispatcher,
) *MarketDataManager {
	return &MarketDataManager{
		client:     client,
		dispatcher: dispatcher,
		refCount:   make(map[string]int),
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
	}
}

func (m *MarketDataManager) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-m.client.Messages():
			candle, key, ok := parseKline(msg)
			if !ok {
				continue
			}
			m.dispatcher.Dispatch(key, candle)
		}
	}
}
