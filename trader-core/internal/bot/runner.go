package bot

import (
	"context"
	"log"

	"trader-core/internal/db"
	"trader-core/internal/db/models"
	"trader-core/internal/engine"

	"github.com/shopspring/decimal"
)

const (
	priceScale    = int64(1e8)
	quantityScale = int64(1e6)
	feeScale      = int64(1e8)
)

func ToInt64(value decimal.Decimal, scale int64) int64 {
	return value.Mul(decimal.NewFromInt(scale)).IntPart()
}

func FromInt64(value int64, scale int64) decimal.Decimal {
	return decimal.NewFromInt(value).Div(decimal.NewFromInt(scale))
}

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

			order := engine.Order{
				BotID:  b.ID,
				Symbol: b.Symbol,
				Side:   side,
				Price:  decimal.NewFromFloat(candle.Close),
				Qty:    decimal.RequireFromString("0.001"),
			}

			fill, err := b.Engine.ExecuteTrade(order)
			if err != nil {
				log.Println("trade error:", err)
				continue
			}

			// scale to int64 for storage
			trade := models.Trade{
				BotID:       fill.BotID,
				Symbol:      fill.Symbol,
				Side:        string(fill.Side),
				PriceInt:    ToInt64(fill.Price, priceScale),
				QuantityInt: ToInt64(fill.Qty, quantityScale),
				FeeInt:      ToInt64(fill.Fee, feeScale),
				FeeAsset:    "USDT",
				Exchange:    "binance",
				Timestamp:   fill.Time,
			}
			if err := db.DB.Create(&trade).Error; err != nil {
				log.Println("Failed to insert trade into DB")
				return
			}
			log.Printf(
				"Bot %s executed %s %s %s @ %s (fee %s)\n",
				fill.BotID,
				fill.Symbol,
				fill.Side,
				fill.Qty.String(),
				fill.Price.String(),
				fill.Fee.String(),
			)
		}
	}
}
