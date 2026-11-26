package bot

import (
	"context"
	"fmt"
	"log"
	"time"

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

			// --- risk management ---
			if b.Risk != nil {
				now := time.Now()
				snap := engine.AccountSnapshot{}
				if as, ok := b.Engine.(engine.AccountSnapshotter); ok {
					snap = as.Snapshot()
				}
				equity := snap.Balance.Add(snap.Positions[b.Symbol()].Mul(candle.Close))

				if kill := b.Risk.TrackEquity(equity, now); kill != "" {
					b.Logger.Warn("kill switch tripped: %s — stopping bot", kill)
					if b.OnKillSwitch != nil {
						go b.OnKillSwitch(kill)
					}
					continue
				}
				if reason, blocked := b.Risk.Blocked(now); blocked && order.Signal == strategies.Buy {
					b.Logger.Info("buy blocked: %s", reason)
					continue
				}
				if order.Signal == strategies.Buy {
					if capped, wasCapped := b.Risk.CapQty(order.Qty, order.Price, snap.Balance); wasCapped {
						b.Logger.Info("position capped: %s -> %s (max_position_pct)", order.Qty.String(), capped.String())
						order.Qty = capped
					}
				}
			}

			f, err := b.Engine.ExecuteFill(order)
			if err != nil {
				log.Println("fill error:", err)
				continue
			}

			if err := persistFill(b, f); err != nil {
				// don't kill the runner — the fill still happened locally; log and carry on
				log.Println("failed to insert fill into DB:", err)
				b.Logger.Warn("failed to persist fill: %v", err)
			}

			if b.Risk != nil {
				b.Risk.NoteFill(f.Signal, f.Pnl, time.Now())
			}

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

// persistFill writes a fill to the DB and broadcasts it to SSE subscribers.
// A nil db.DB (unit tests) skips persistence.
func persistFill(b *Bot, f *engine.Fill) error {
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

	if db.DB != nil {
		if err := db.DB.Create(&fill).Error; err != nil {
			return err
		}
	}

	// broadcast to SSE subscribers
	b.FillBroadcaster.Publish(dto.FIllToDTO(&fill))
	return nil
}

// Flatten sells every open position of the bot's account at the last candle
// close (paper fills locally; live places real market SELLs via the placer).
// Returns the symbols that were flattened.
func Flatten(b *Bot) ([]string, error) {
	as, ok := b.Engine.(engine.AccountSnapshotter)
	if !ok {
		return nil, fmt.Errorf("bot engine has no account ledger")
	}
	snap := as.Snapshot()
	if len(b.Candles) == 0 {
		return nil, fmt.Errorf("no candles loaded, cannot price positions")
	}
	price := b.Candles[len(b.Candles)-1].Close

	var flattened []string
	for sym, qty := range snap.Positions {
		if qty.LessThanOrEqual(decimal.Zero) {
			continue
		}
		f, err := b.Engine.ExecuteFill(engine.Order{
			BotID:  b.ID,
			Symbol: sym,
			Signal: strategies.Sell,
			Price:  price,
			Qty:    qty,
		})
		if err != nil {
			return flattened, fmt.Errorf("flatten %s: %w", sym, err)
		}
		if err := persistFill(b, f); err != nil {
			b.Logger.Warn("failed to persist flatten fill: %v", err)
		}
		b.Logger.Warn("flattened %s %s @ %s", sym, f.Qty.String(), f.Price.String())
		flattened = append(flattened, sym)
	}
	return flattened, nil
}
