package db

import (
	"encoding/json"
	"log"
	"os"
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
	{Slug: "sma_cross", DisplayName: "SMA Crossover", DefaultConfig: map[string]any{}},
	{Slug: "ml_example", DisplayName: "ML Example (HTTP)", DefaultConfig: map[string]any{}},
	{Slug: "sentiment", DisplayName: "Sentiment (HTTP)", DefaultConfig: sentimentDefaultConfig()},
	// {Name: "smart_dca", DisplayName: "Smart DCA", DefaultConfig: map[string]any{"fixed_spend": "10.00"}},
	// {Name: "macd", DisplayName: "MACD", DefaultConfig: map[string]any{"fast": 12, "slow": 26, "signal": 9}},
}

// sentimentDefaultConfig lets deployments point the seeded sentiment strategy
// at the in-network service (compose sets SENTIMENT_URL, e.g.
// http://sentiment:8002/predict); empty = schema default (127.0.0.1:8002).
func sentimentDefaultConfig() map[string]any {
	cfg := map[string]any{}
	if u := os.Getenv("SENTIMENT_URL"); u != "" {
		cfg["url"] = u
	}
	return cfg
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
