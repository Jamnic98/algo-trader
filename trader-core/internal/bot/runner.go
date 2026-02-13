package bot

import (
	"context"
	"log"
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
			if len(b.Candles) > b.MaxLookback {
				b.Candles = b.Candles[len(b.Candles)-b.MaxLookback:]
			}

			// need enough history
			if len(b.Candles) < 20 {
				continue
			}

			side := b.Strategy.OnCandles(b.Candles)
			if side == NONE {
				continue
			}

			err := b.Engine.ExecuteTrade(
				b.Symbol,
				string(side),
				candle.Close,
				0.001,
			)
			if err != nil {
				log.Println("trade error:", err)
			}
		}
	}
}
