package bot

import (
	"context"
	"log"

	"trader-core/internal/engine"
)

func RunBotStrategy(ctx context.Context, b *Bot) {
	for {
		select {
		case <-ctx.Done():
			b.Status = "stopped"
			return

		case candle := <-b.CandleCh:
			b.Candles = append(b.Candles, candle)

			// cap lookback
			if len(b.Candles) > b.MaxCandles {
				b.Candles = b.Candles[len(b.Candles)-b.MaxCandles:]
			}

			// need enough history
			if len(b.Candles) < 5 {
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
			} else {
				// TODO: persist to DB or do something with the fill
				// Replace LogTrade with a simple log statement
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
}
