package dto

import (
	"trader-core/internal/db/models"
)

// StrategyConfigDTO — the DefaultConfig blob
type StrategyConfigDTO struct {
	Name          string              `json:"name"`
	DisplayName   string              `json:"display_name"`
	Params        map[string]any      `json:"params,omitempty"`
	MakerFee      string              `json:"maker_fee"`
	TakerFee      string              `json:"taker_fee"`
	SubStrategies []StrategyConfigDTO `json:"sub_strategies,omitempty"`
}

func StrategyConfigToDTO(sc models.Strategy) (StrategyConfigDTO, error) {
	var dto StrategyConfigDTO
	// if err := json.Unmarshal(sc.DefaultConfig, &dto); err != nil {
	// 	return StrategyConfigDTO{}, fmt.Errorf("parsing strategy config: %w", err)
	// }
	return dto, nil
}
