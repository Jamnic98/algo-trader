package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"trader-core/internal/binance"
	"trader-core/internal/bot"
	"trader-core/internal/db/models"
	"trader-core/internal/engine"
	"trader-core/setup"

	"github.com/google/uuid"
)

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

func main() {
	// logger setup
	f, err := os.OpenFile("bot.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Tell the logger to write to the file
	log.SetOutput(f)

	log.Println("Bot started")
	log.Println("Candle received")

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := setup.GetConfig()
	setup.InitDatabase(cfg)

	server := setup.InitServer()
	wg.Go(func() {
		if err := server.Run(":" + cfg.Port); err != nil {
			log.Fatal("Failed to start server:", err)
		}
	})

	binanceClient := binance.NewClient(ctx, "wss://stream.binance.com:443/ws")
	wg.Go(func() {
		if err := binanceClient.Run(); err != nil {
			log.Fatal(err)
		}
	})

	binanceClient.Subscribe("btcusdt@kline_1m")

	b := bot.Bot{
		ID:          uuid.New().String(),
		Symbol:      "BTCUSDT",
		Interval:    "1m",
		Status:      "running",
		Started:     time.Now(),
		Engine:      engine.NewPaperExecution(10000, 0.001),
		Strategy:    bot.SimpleStrategy{},
		MaxLookback: 100,
		CandleCh:    make(chan models.Candle, 100),
	}

	wg.Add(1)
	go bot.RunBotStrategy(ctx, &b)

	for msg := range binanceClient.Messages() {
		var evt KlineEvent
		if err := json.Unmarshal(msg, &evt); err != nil {
			log.Println("unmarshal error:", err)
			continue
		}

		if evt.K.IsClosed {
			// Convert numbers to float64 if needed
			open, _ := evt.K.Open.Float64()
			high, _ := evt.K.High.Float64()
			low, _ := evt.K.Low.Float64()
			closePrice, _ := evt.K.Close.Float64()
			volume, _ := evt.K.Volume.Float64()

			log.Printf("Candle closed for %s %s: O=%f H=%f L=%f C=%f V=%f\n",
				evt.Symbol, evt.K.Interval, open, high, low, closePrice, volume)
			log.Println()

			candle := models.Candle{
				Open:   open,
				High:   high,
				Low:    low,
				Close:  closePrice,
				Volume: volume,
			}

			// only send candles the bot cares about
			if b.Symbol == evt.Symbol && b.Interval == "1m" {
				select {
				case b.CandleCh <- candle:
				default:
					log.Println("bot candle buffer full, dropping candle")
				}
			}
		}
	}

	wg.Wait()
	<-ctx.Done()
}
