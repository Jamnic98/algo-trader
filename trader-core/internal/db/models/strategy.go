package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type StrategyNode struct {
	StrategyID uint    `json:"strategy_id"`
	Weight     float64 `json:"weight"`
}

type Strategy struct {
	gorm.Model
	Name          string         `gorm:"uniqueIndex;not null"`
	DisplayName   string         `gorm:"not null"`
	IsComposite   bool           `gorm:"not null;default:false"`
	DefaultConfig datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'"`
	// For composite strategies — weighted sub-strategy references
	// Stored as JSONB: [{"strategy_id": 1, "weight": 0.5}, ...]
	Nodes datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'"`
}
