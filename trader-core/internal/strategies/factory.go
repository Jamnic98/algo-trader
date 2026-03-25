package strategies

import (
	"encoding/json"
	"fmt"
	"trader-core/internal/common"
	"trader-core/internal/db/models"
	"trader-core/internal/dto"

	"github.com/shopspring/decimal"
)

func NewFromConfig(cfg models.Strategy, logger Logger) (Strategy, error) {
	// first unmarshal the full config blob into the DTO
	var stratDTO dto.StrategyConfigDTO
	if err := json.Unmarshal(cfg.DefaultConfig, &stratDTO); err != nil {
		return nil, fmt.Errorf("parsing strategy config: %w", err)
	}

	// then unmarshal params into the typed struct for each strategy
	switch cfg.Name {
	case "simple":
		var params SimpleParams
		if err := unmarshalParams(stratDTO.Params, &params); err != nil {
			return nil, fmt.Errorf("invalid params for simple: %w", err)
		}
		return NewSimpleStrategy(params), nil

	// case "simpleDca":
	// 	var params SimpleDCAParams
	// 	if err := unmarshalParams(stratDTO.Params, &params); err != nil {
	// 		return nil, fmt.Errorf("invalid params for simpleDca: %w", err)
	// 	}
	// 	return NewSimpleDCAStrategy(params, logger), nil

	// case "smartDca":
	// 	var params SmartDCAParams
	// 	if err := unmarshalParams(stratDTO.Params, &params); err != nil {
	// 		return nil, fmt.Errorf("invalid params for smartDca: %w", err)
	// 	}
	// 	return NewSmartDCAStrategy(params, logger), nil

	default:
		return nil, fmt.Errorf("unknown strategy: %q", cfg.Name)
	}
}

// unmarshalParams round-trips map[string]any → JSON → typed struct
func unmarshalParams(params map[string]any, dst any) error {
	b, err := json.Marshal(params)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}

func DefaultFees() common.FeeConfig {
	return common.FeeConfig{
		MakerFee: decimal.NewFromFloat(0.001),
		TakerFee: decimal.NewFromFloat(0.001),
	}
}
