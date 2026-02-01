package models

type Candle struct {
    ID        uint      `gorm:"primaryKey"`
    Symbol    string
    Open      float64
    Close     float64
    // Timestamp time.Time
}
