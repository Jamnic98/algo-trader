package setup

import (
	"trader-core/internal/bot"
	"trader-core/internal/db"
	"trader-core/internal/db/models"
	_ "trader-core/internal/strategies"

	"gorm.io/gorm"
)

func InitDatabase(cfg Config) *gorm.DB {
	// Connect to Postgres
	db.ConnectPostgres(cfg.Dsn)

	// Auto-migrate tables
	db.Migrate(&bot.BotConfig{}, &models.Trade{}, &models.Strategy{})

	// Seed db
	db.SeedStrategies(db.DB)

	return db.DB
}
