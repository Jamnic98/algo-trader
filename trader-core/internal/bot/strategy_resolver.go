package bot

import (
	"encoding/json"
	"fmt"
	"trader-core/internal/db/models"
	"trader-core/internal/strategies"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func ResolveStrategy(db *gorm.DB, s models.Strategy) (strategies.Strategy, error) {
	if !s.IsComposite {
		return strategies.Build(s.Slug, s.DefaultConfig)
	}
	return resolveComposite(db, s)
}

// resolveComposite builds a WeightedComposer from the strategy's DefaultConfig
// ({"threshold": ..., "nodes": [{"strategy_slug", "weight", "params"}]}),
// falling back to the legacy Nodes column if DefaultConfig has no nodes.
func resolveComposite(db *gorm.DB, s models.Strategy) (strategies.Strategy, error) {
	var cfg strategies.CompositeConfig
	if len(s.DefaultConfig) > 0 {
		if err := json.Unmarshal(s.DefaultConfig, &cfg); err != nil {
			return nil, fmt.Errorf("parsing composite config: %w", err)
		}
	}

	if len(cfg.Nodes) == 0 && len(s.Nodes) > 0 {
		var legacy []models.StrategyNode
		if err := json.Unmarshal(s.Nodes, &legacy); err != nil {
			return nil, fmt.Errorf("parsing composite nodes: %w", err)
		}
		for _, n := range legacy {
			cfg.Nodes = append(cfg.Nodes, strategies.CompositeNode{
				StrategySlug: n.StrategySlug,
				Weight:       n.Weight,
			})
		}
	}

	if len(cfg.Nodes) == 0 {
		return nil, fmt.Errorf("composite strategy %q has no nodes", s.Slug)
	}

	threshold := cfg.Threshold
	if threshold <= 0 {
		threshold = 0.6
	}

	composer := &strategies.WeightedComposer{Threshold: threshold}
	for _, node := range cfg.Nodes {
		subStrategy, err := resolveBySlug(db, node.StrategySlug, node.Params) // recursive — handles nested composites
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

// resolveBySlug loads a strategy row by slug and builds it, merging any
// per-node param overrides on top of the sub-strategy's default config.
func resolveBySlug(db *gorm.DB, slug string, overrides json.RawMessage) (strategies.Strategy, error) {
	var s models.Strategy
	if err := db.First(&s, "slug = ?", slug).Error; err != nil {
		return nil, fmt.Errorf("loading sub-strategy %q: %w", slug, err)
	}

	if !s.IsComposite {
		cfg := s.DefaultConfig
		if len(overrides) > 0 {
			merged, err := mergeParams(json.RawMessage(cfg), overrides)
			if err != nil {
				return nil, fmt.Errorf("merging params for %q: %w", slug, err)
			}
			cfg = datatypes.JSON(merged)
		}
		return strategies.Build(s.Slug, json.RawMessage(cfg))
	}

	return resolveComposite(db, s)
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
