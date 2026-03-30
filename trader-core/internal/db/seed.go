package db

import (
	"encoding/json"
	"log"
	"trader-core/internal/db/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type primitiveSeed struct {
	Slug          string
	DisplayName   string
	DefaultConfig map[string]any
}

var primitiveStrategies = []primitiveSeed{
	{Slug: "simple", DisplayName: "Simple", DefaultConfig: map[string]any{}},
	{Slug: "simple_dca", DisplayName: "Simple DCA", DefaultConfig: map[string]any{}},
	// {Name: "smart_dca", DisplayName: "Smart DCA", DefaultConfig: map[string]any{"fixed_spend": "10.00"}},
	// {Name: "macd", DisplayName: "MACD", DefaultConfig: map[string]any{"fast": 12, "slow": 26, "signal": 9}},
}

func SeedStrategies(db *gorm.DB) {
	for _, s := range primitiveStrategies {
		cfg, _ := json.Marshal(s.DefaultConfig)
		record := models.Strategy{
			Slug:          s.Slug,
			DisplayName:   s.DisplayName,
			IsComposite:   false,
			DefaultConfig: cfg,
			Nodes:         []byte("[]"),
		}
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "slug"}},
			DoUpdates: clause.AssignmentColumns([]string{"display_name", "default_config"}),
		}).Create(&record).Error; err != nil {
			log.Printf("seed strategy %q: %v", s.Slug, err)
		}
	}
}
