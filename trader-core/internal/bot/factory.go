package bot

import (
	"fmt"
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

	strategy, err := strategies.NewStrategy(cfg.StrategyConfig.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get strategy: %w", err)
	}

	b := &Bot{
		BotConfig: cfg,
		Status:    BotCreated,

		Engine:   f.Engine(),
		Logger:   NewBotLogger(cfg.ID),
		Strategy: strategy,
		CandleCh: make(chan models.Candle, 100),

		TradeBroadcaster:  &Broadcaster[dto.TradeDTO]{},
		CandleBroadcaster: &Broadcaster[models.Candle]{},
		TickBroadcaster:   &Broadcaster[models.Candle]{},
	}

	return b, nil
}
