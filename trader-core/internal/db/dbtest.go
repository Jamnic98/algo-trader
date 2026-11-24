package db

import (
	"fmt"
	"net/url"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ConnectTestDatabase creates (idempotently) a dedicated database derived from
// baseDSN and points the package-global DB at it. Lets parallel `go test`
// packages each run AutoMigrate without racing on CREATE TABLE.
func ConnectTestDatabase(baseDSN, name string) error {
	u, err := url.Parse(baseDSN)
	if err != nil {
		return fmt.Errorf("parse TEST_DB_DSN: %w", err)
	}

	admin := *u
	admin.Path = "/postgres"
	adminDB, err := gorm.Open(postgres.Open(admin.String()), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("admin connect: %w", err)
	}

	var exists bool
	if err := adminDB.Raw(
		"SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = ?)", name,
	).Scan(&exists).Error; err != nil {
		return fmt.Errorf("check database: %w", err)
	}
	if !exists {
		if err := adminDB.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, name)).Error; err != nil {
			return fmt.Errorf("create database %q: %w", name, err)
		}
	}

	u.Path = "/" + name
	ConnectPostgres(u.String())
	return nil
}
