package models

import (
	"time"

	"gorm.io/datatypes"
)

type StrategyNode struct {
	StrategySlug string  `json:"strategy_slug"`
	Weight       float64 `json:"weight"`
}

type Strategy struct {
	Slug          string         `gorm:"primaryKey"`
	DisplayName   string         `gorm:"not null"`
	IsComposite   bool           `gorm:"not null;default:false"`
	DefaultConfig datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'"`
	Nodes         datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
