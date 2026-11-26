package bot

import (
	"testing"

	"trader-core/internal/db/models"

	"github.com/shopspring/decimal"
)

func TestDrainCandlesDropsCoveredKeepsNewer(t *testing.T) {
	ch := make(chan models.Candle, 10)
	mk := func(closeTime int64) models.Candle {
		return models.Candle{CloseTime: closeTime, Close: decimal.NewFromInt(closeTime)}
	}
	ch <- mk(100) // covered by history
	ch <- mk(200) // covered
	ch <- mk(300) // newer than lastClose=200 -> kept
	ch <- mk(400) // kept

	drainCandles(ch, 200)

	if len(ch) != 2 {
		t.Fatalf("len = %d, want 2", len(ch))
	}
	if got := (<-ch).CloseTime; got != 300 {
		t.Fatalf("first kept = %d, want 300", got)
	}
	if got := (<-ch).CloseTime; got != 400 {
		t.Fatalf("second kept = %d, want 400", got)
	}
}

func TestDrainCandlesEmpty(t *testing.T) {
	ch := make(chan models.Candle, 10)
	drainCandles(ch, 999) // must not block or panic
	if len(ch) != 0 {
		t.Fatalf("len = %d, want 0", len(ch))
	}
}
