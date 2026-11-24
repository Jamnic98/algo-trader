package models

import "time"

// Account is a persisted trading account definition (paper accounts keep
// their ledger balance here; the live account row is refreshed from Binance).
type Account struct {
	Name      string `gorm:"primaryKey"     json:"name"`
	Type      string `gorm:"not null"       json:"type"`   // "paper" | "live"
	Balance   string `gorm:"not null;default:'0'" json:"balance"` // quote currency, decimal string
	Fee       string `gorm:"not null;default:'0.001'" json:"fee"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
