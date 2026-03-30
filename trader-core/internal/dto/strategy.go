package dto

import (
	"encoding/json"
	"time"
	"trader-core/internal/db/models"
)

type StrategyDTO struct {
	Slug        string          `json:"slug"`
	DisplayName string          `json:"display_name"`
	IsComposite bool            `json:"is_composite"`
	CreatedAt   string          `json:"created_at"`
	Config      json.RawMessage `json:"config,omitempty"`
}

func StrategyToDTO(s *models.Strategy) StrategyDTO {
	dto := StrategyDTO{
		Slug:        s.Slug,
		DisplayName: s.DisplayName,
		IsComposite: s.IsComposite,
		CreatedAt:   s.CreatedAt.Format(time.RFC3339),
	}
	if s.IsComposite && len(s.DefaultConfig) > 0 {
		dto.Config = json.RawMessage(s.DefaultConfig)
	}
	return dto
}
