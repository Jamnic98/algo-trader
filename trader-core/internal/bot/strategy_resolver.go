package bot

import (
	"encoding/json"
	"fmt"
	"trader-core/internal/db/models"
	"trader-core/internal/strategies"

	"gorm.io/gorm"
)

func ResolveStrategy(db *gorm.DB, s models.Strategy) (strategies.Strategy, error) {
	if !s.IsComposite {
		return strategies.Build(s.Name, s.DefaultConfig)
	}

	var nodes []models.StrategyNode
	if err := json.Unmarshal(s.Nodes, &nodes); err != nil {
		return nil, fmt.Errorf("parsing composite nodes: %w", err)
	}

	composer := &strategies.WeightedComposer{Threshold: 0.6}
	for _, node := range nodes {
		var sub models.Strategy
		if err := db.First(&sub, node.StrategyID).Error; err != nil {
			return nil, fmt.Errorf("loading sub-strategy %d: %w", node.StrategyID, err)
		}
		subStrategy, err := ResolveStrategy(db, sub) // recursive — handles nested composites
		if err != nil {
			return nil, err
		}
		composer.Nodes = append(composer.Nodes, strategies.ComposerNode{
			Strategy: subStrategy,
			Weight:   node.Weight,
		})
	}

	return composer, nil
}

// mergeParams merges override params on top of base params JSON
func mergeParams(base, overrides json.RawMessage) (json.RawMessage, error) {
	if len(overrides) == 0 {
		return base, nil
	}

	baseMap := map[string]any{}
	if len(base) > 0 {
		if err := json.Unmarshal(base, &baseMap); err != nil {
			return nil, err
		}
	}

	overrideMap := map[string]any{}
	if err := json.Unmarshal(overrides, &overrideMap); err != nil {
		return nil, err
	}

	for k, v := range overrideMap {
		baseMap[k] = v
	}

	return json.Marshal(baseMap)
}
