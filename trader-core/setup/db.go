package setup

import (
	"trader-core/internal/db"
	"trader-core/internal/db/models"
)

func InitDatabase(cfg Config) {
	// Connect to Postgres
	db.ConnectPostgres(cfg.Dsn)

	// Auto-migrate tables
	db.Migrate(&models.Trade{})
}
