package strategies

import (
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
)

type ComposerNode struct {
	Strategy Strategy
	Weight   float64
}

type WeightedComposer struct {
	Nodes     []ComposerNode
	Threshold float64
}

func (w *WeightedComposer) OnCandle(ctx CandleContext) Decision {
	var buyScore, sellScore, totalWeight float64

	for _, node := range w.Nodes {
		d := node.Strategy.OnCandle(ctx)
		totalWeight += node.Weight
		switch d.Signal {
		case Buy:
			buyScore += node.Weight
		case Sell:
			sellScore += node.Weight
		}
	}

	if totalWeight == 0 {
		return Decision{Signal: Hold}
	}

	buyScore /= totalWeight
	sellScore /= totalWeight

	if buyScore >= w.Threshold {
		return Decision{Signal: Buy}
	}
	if sellScore >= w.Threshold {
		return Decision{Signal: Sell}
	}
	return Decision{
		Signal: Hold,
		Reason: fmt.Sprintf("buy=%.2f sell=%.2f threshold=%.2f", buyScore, sellScore, w.Threshold),
	}
}

// CompositeConfig is the shape stored in DefaultConfig for composite strategies
type CompositeConfig struct {
	Threshold float64         `json:"threshold"`
	Nodes     []CompositeNode `json:"nodes"`
}

type CompositeNode struct {
	StrategyID uint            `json:"strategy_id"`
	Weight     float64         `json:"weight"`
	Params     json.RawMessage `json:"params,omitempty"`
}

// keep decimal import happy
var _ = decimal.Zero
