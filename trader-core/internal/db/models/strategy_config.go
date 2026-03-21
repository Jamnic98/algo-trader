package models

import (
	"github.com/shopspring/decimal"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type StrategyConfig struct {
	gorm.Model
	BotID       string          `json:"bot_id"       gorm:"not null;index"`
	Name        string          `json:"name"         gorm:"not null"`
	DisplayName string          `json:"display_name" gorm:"not null"`
	Params      datatypes.JSON  `json:"params"       gorm:"type:jsonb;default:'{}'"`
	MakerFee    decimal.Decimal `json:"maker_fee"   gorm:"type:numeric(10,8);default:0.001"`
	TakerFee    decimal.Decimal `json:"taker_fee"   gorm:"type:numeric(10,8);default:0.001"`
}
