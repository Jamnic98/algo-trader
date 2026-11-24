package db

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectPostgres(dsn string) {
    var err error
    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect Postgres:", err)
    }
}

func Migrate(models ...any) {
    if err := DB.AutoMigrate(models...); err != nil {
        log.Fatal("Migration failed:", err)
    }
}
