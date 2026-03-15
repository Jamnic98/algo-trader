package bot

import (
	"trader-core/internal/db/models"
	"trader-core/internal/dto"
	"trader-core/internal/engine"
	"trader-core/internal/strategies"

	"github.com/google/uuid"
)

type BotFactory struct {
	Account engine.Account
	Engine  func() engine.ExecutionEngine
}

func (f *BotFactory) NewPaperBot(cfg BotConfig) (*Bot, error) {
	if cfg.ID == "" {
		cfg.ID = uuid.New().String()
	}

	b := &Bot{
		BotConfig: cfg,
		Status:    BotCreated,

		Engine:   f.Engine(),
		Logger:   NewBotLogger(cfg.ID),
		Strategy: &strategies.SimpleStrategy{},
		CandleCh: make(chan models.Candle, 100),

		TradeBroadcaster:  &Broadcaster[dto.TradeDTO]{},
		CandleBroadcaster: &Broadcaster[models.Candle]{},
	}

	return b, nil
}
