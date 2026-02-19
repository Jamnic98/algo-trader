package bot

import (
	"context"
	"log"
	"time"

	"trader-core/internal/db"
	"trader-core/internal/db/models"
	"trader-core/internal/engine"
)

func RunBotStrategy(ctx context.Context, b *Bot) {
	for {
		select {
		case <-ctx.Done():
			b.Status = BotAttached
			return

		case candle, ok := <-b.CandleCh:
			if !ok {
				b.Status = BotAttached
				return
			}
			b.Candles = append(b.Candles, candle)

			// cap lookback
			if len(b.Candles) > b.MaxCandles {
				b.Candles = b.Candles[len(b.Candles)-b.MaxCandles:]
			}

			// need enough history
			if len(b.Candles) < 2 {
				continue
			}

			side := b.Strategy.OnCandles(b.Candles)
			if side == engine.NONE {
				continue
			}

			fill, err := b.Engine.ExecuteTrade(
				engine.Order{
					BotID:  b.ID,         // bot ID
					Symbol: b.Symbol,     // symbol
					Side:   side,         // side of type engine.Side
					Price:  candle.Close, // price
					Qty:    0.001,        // qty
				},
			)
			if err != nil {
				log.Println("trade error:", err)
				continue
			}
			trade := models.Trade{
				BotID:     fill.BotID,
				Symbol:    fill.Symbol,
				Price:     fill.Price,
				Fee:       fill.Fee,
				FeeAsset:  "USDT",
				Side:      string(fill.Side),
				Quantity:  fill.Qty,
				Exchange:  "binance",
				Timestamp: time.Now(),
			}
			if err := db.DB.Create(&trade).Error; err != nil {
				log.Println("Failed to insert trade into DB")
				return
			}
			log.Printf(
				"Bot %s executed %s %.4f %s @ %.2f (fee %.6f)\n",
				fill.BotID,
				fill.Side,
				fill.Qty,
				fill.Symbol,
				fill.Price,
				fill.Fee,
			)
		}
	}
}
