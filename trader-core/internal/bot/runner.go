package bot

import (
	"context"
	"log"

	"trader-core/internal/db"
	"trader-core/internal/db/models"
	"trader-core/internal/dto"
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
			return

		case candle, ok := <-b.CandleCh:
			if !ok {
				return
			}

			// always update candle buffer regardless of status
			b.Candles = append(b.Candles, candle)
			if len(b.Candles) > b.MaxCandles {
				b.Candles = b.Candles[len(b.Candles)-b.MaxCandles:]
			}
			b.CandleBroadcaster.Publish(candle)

			// only trade when running
			if b.Status != BotRunning {
				continue
			}

			side := b.Strategy.OnCandles(b.Candles)
			b.Logger.Info("strategy signal: %s", side)
			if side == engine.NONE {
				continue
			}

			order := engine.Order{
				BotID:  b.ID,
				Symbol: b.Symbol(),
				Side:   side,
				Price:  candle.Close,
				Qty:    b.Quantity,
			}

			fill, err := b.Engine.ExecuteTrade(order)
			if err != nil {
				log.Println("trade error:", err)
				continue
			}

			trade := models.Trade{
				BotID:       fill.BotID,
				Symbol:      fill.Symbol,
				Base:        b.Base,
				Quote:       b.Quote,
				Side:        string(fill.Side),
				PriceInt:    ToInt64(fill.Price, priceScale),
				QuantityInt: ToInt64(fill.Qty, quantityScale),
				FeeInt:      ToInt64(fill.Fee, feeScale),
				FeeAsset:    b.Quote,
				// TODO: remove from hardcoding
				Exchange:  "binance",
				Timestamp: fill.Time,
			}

			if err := db.DB.Create(&trade).Error; err != nil {
				log.Println("Failed to insert trade into DB")
				return
			}

			// broadcast to SSE subscribers
			b.TradeBroadcaster.Publish(dto.TradeToDTO(&trade))

			b.Logger.Info(
				"%s %s %s @ %s (fee %s)\n",
				fill.Side,
				fill.Symbol,
				fill.Qty.String(),
				fill.Price.String(),
				fill.Fee.String(),
			)
		}
	}
}
