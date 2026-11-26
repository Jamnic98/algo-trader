package models

import (
	"time"

	"gorm.io/datatypes"
)

// Account is a persisted trading account definition (paper accounts keep
// their ledger balance AND open positions here; the live account row is
// refreshed from the exchange).
type Account struct {
	Name      string         `gorm:"primaryKey"     json:"name"`
	Type      string         `gorm:"not null"       json:"type"`   // "paper" | "live"
	Balance   string         `gorm:"not null;default:'0'" json:"balance"` // quote currency, decimal string
	Fee       string         `gorm:"not null;default:'0.001'" json:"fee"`
	Positions datatypes.JSON `gorm:"type:jsonb"       json:"positions"` // symbol -> qty (decimal string), paper accounts
	CreatedAt time.Time
	UpdatedAt time.Time
}
