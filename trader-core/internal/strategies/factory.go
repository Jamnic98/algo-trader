package strategies

import (
	"trader-core/internal/common"
	"trader-core/internal/db/models"

	"github.com/shopspring/decimal"
)

// NewFromConfig builds a strategy from a DB row. The registry is populated
// from shared/strategies.json (see definitions.go), so this just delegates.
func NewFromConfig(cfg models.Strategy, logger Logger) (Strategy, error) {
	return Build(cfg.Slug, cfg.DefaultConfig)
}

func DefaultFees() common.FeeConfig {
	return common.FeeConfig{
		MakerFee: decimal.NewFromFloat(0.001),
		TakerFee: decimal.NewFromFloat(0.001),
	}
}
