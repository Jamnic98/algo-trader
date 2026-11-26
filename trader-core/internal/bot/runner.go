package bot

import (
	"context"
	"log"

	"trader-core/internal/db"
	"trader-core/internal/db/models"
	"trader-core/internal/dto"
	"trader-core/internal/engine"
	"trader-core/internal/strategies"

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

			b.Candles = append(b.Candles, candle)
			if len(b.Candles) > b.MaxCandles {
				b.Candles = b.Candles[len(b.Candles)-b.MaxCandles:]
			}
			b.CandleBroadcaster.Publish(candle)

			if b.Status != BotRunning {
				continue
			}

			feeRate := decimal.NewFromFloat32(0.001)
			if fr, ok := b.Engine.(engine.FeeRater); ok {
				feeRate = fr.FeeRate()
			}
			decision := b.ActiveStrategy.OnCandle(strategies.CandleContext{
				Candle:  candle,
				FeeRate: feeRate,
			})

			if decision.Reason != "" {
				b.Logger.Info("decision: %s — %s", decision.Signal, decision.Reason)
			} else {
				b.Logger.Info("decision: %s", decision.Signal)
			}

			if decision.Signal == strategies.Hold {
				continue
			}

			var quantity decimal.Decimal
			if decision.Quantity != nil {
				quantity = *decision.Quantity
			} else {
				var err error
				quantity, err = decimal.NewFromString(b.Quantity)
				if err != nil {
					log.Println("invalid quantity:", err)
					continue
				}
			}

			order := engine.Order{
				BotID:  b.ID,
				Symbol: b.Symbol(),
				Signal: strategies.Signal(decision.Signal),
				Price:  candle.Close,
				Qty:    quantity,
			}

			f, err := b.Engine.ExecuteFill(order)
			if err != nil {
				log.Println("fill error:", err)
				continue
			}

			fill := models.Fill{
				BotID:       f.BotID,
				Symbol:      f.Symbol,
				Base:        b.Base,
				Quote:       b.Quote,
				Side:        string(f.Signal),
				PriceInt:    ToInt64(f.Price, priceScale),
				QuantityInt: ToInt64(f.Qty, quantityScale),
				FeeInt:      ToInt64(f.Fee, feeScale),
				FeeAsset:    b.Quote,
				Exchange:    b.Exchange,
				Timestamp:   f.Time,
			}

			if err := db.DB.Create(&fill).Error; err != nil {
				// don't kill the runner — the fill still happened locally; log and carry on
				log.Println("failed to insert fill into DB:", err)
				b.Logger.Warn("failed to persist fill: %v", err)
			}

			// broadcast to SSE subscribers
			b.FillBroadcaster.Publish(dto.FIllToDTO(&fill))

			b.Logger.Info(
				"%s %s %s @ %s (fee %s)\n",
				f.Signal,
				f.Symbol,
				f.Qty.String(),
				f.Price.String(),
				f.Fee.String(),
			)
		}
	}
}
