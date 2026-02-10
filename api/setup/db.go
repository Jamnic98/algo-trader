package setup

import (
	"api/internal/db"
	"api/internal/models"
)


func InitDatabase(cfg Config) {
  // Connect to Postgres
  db.ConnectPostgres(cfg.Dsn)

  // Auto-migrate tables
  db.Migrate(&models.Bot{})
  db.Migrate(&models.Candle{})
  db.Migrate(&models.Trade{})
}
