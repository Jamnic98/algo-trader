package binance

import (
	"log"
	"sync"

	"trader-core/internal/bot"
	"trader-core/internal/db/models"
)

type Dispatcher struct {
	subscriptions map[string][]*bot.Bot
	mu            sync.RWMutex
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		subscriptions: make(map[string][]*bot.Bot),
	}
}

// Subscribe a bot to a feed
func (d *Dispatcher) Subscribe(symbol string, interval bot.Interval, b *bot.Bot) {
	key := symbol + "_" + interval.String()
	d.mu.Lock()
	defer d.mu.Unlock()
	d.subscriptions[key] = append(d.subscriptions[key], b)
}

// Unsubscribe a bot
func (d *Dispatcher) Unsubscribe(symbol string, interval bot.Interval, b *bot.Bot) {
	key := symbol + "_" + interval.String()
	d.mu.Lock()
	defer d.mu.Unlock()
	bots := d.subscriptions[key]
	for i, bot := range bots {
		if bot == b {
			d.subscriptions[key] = append(bots[:i], bots[i+1:]...)
			break
		}
	}
}

// Dispatch a candle to all bots subscribed to that feed
func (d *Dispatcher) Dispatch(key string, candle models.Candle) {
	d.mu.RLock()
	bots := d.subscriptions[key]
	d.mu.RUnlock()

	for _, b := range bots {
		select {
		case b.CandleCh <- candle:
			log.Printf("Dispatched candle to bot %s", b.ID)
		default:
			log.Printf("Dropped candle for bot %s (channel full)", b.ID)
		}
	}
}
