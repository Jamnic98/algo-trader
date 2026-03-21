package strategies

import (
	"encoding/json"
	"fmt"
	"trader-core/internal/common"
	"trader-core/internal/db/models"

	"github.com/shopspring/decimal"
)

func NewFromConfig(cfg models.StrategyConfig, logger Logger) (Strategy, error) {
	switch cfg.Name {
	case "simple":
		var params SimpleParams
		if err := json.Unmarshal(cfg.Params, &params); err != nil {
			return nil, fmt.Errorf("invalid params for simple: %w", err)
		}
		return NewSimpleStrategy(params, logger), nil
	case "simpleDca":
		var params SimpleDCAParams
		if err := json.Unmarshal(cfg.Params, &params); err != nil {
			return nil, fmt.Errorf("invalid params for simpleDca: %w", err)
		}
		return NewSimpleDCAStrategy(params, logger), nil
	case "smartDca":
		var params SmartDCAParams
		if err := json.Unmarshal(cfg.Params, &params); err != nil {
			return nil, fmt.Errorf("invalid params for smartDca: %w", err)
		}
		return NewSmartDCAStrategy(params, logger), nil
	default:
		return nil, fmt.Errorf("unknown strategy: %q", cfg.Name)
	}
}

// DefaultFees returns standard Binance spot fees
func DefaultFees() common.FeeConfig {
	return common.FeeConfig{
		MakerFee: decimal.NewFromFloat(0.001),
		TakerFee: decimal.NewFromFloat(0.001),
	}
}
