package bot

import (
	"trader-core/internal/db/models"
	"trader-core/internal/dto"
	"trader-core/internal/engine"
	"trader-core/internal/strategies"

	"github.com/google/uuid"
)

type BotFactory struct {
	Account     engine.Account
	Engine      func() engine.ExecutionEngine
	LiveAccount *engine.LiveAccount
	Placer      engine.OrderPlacer
}

func (f *BotFactory) newBot(cfg BotConfig, strategy strategies.Strategy, exec engine.ExecutionEngine) (*Bot, error) {
	if cfg.ID == "" {
		cfg.ID = uuid.New().String()
	}

	return &Bot{
		BotConfig: cfg,
		Status:    BotCreated,

		Engine:         exec,
		Logger:         NewBotLogger(cfg.ID),
		ActiveStrategy: strategy,
		CandleCh:       make(chan models.Candle, 100),

		FillBroadcaster:   &Broadcaster[dto.FIllDTO]{},
		CandleBroadcaster: &Broadcaster[models.Candle]{},
		TickBroadcaster:   &Broadcaster[models.Candle]{},
	}, nil
}

func (f *BotFactory) NewPaperBot(cfg BotConfig, strategy strategies.Strategy) (*Bot, error) {
	return f.newBot(cfg, strategy, f.Engine())
}

// NewLiveBot builds a bot whose fills are placed on the exchange via Placer.
func (f *BotFactory) NewLiveBot(cfg BotConfig, strategy strategies.Strategy) (*Bot, error) {
	acc := f.LiveAccount
	if acc == nil {
		acc = engine.NewLiveAccount("0", "0.001")
	}
	return f.newBot(cfg, strategy, engine.NewLiveExecutor(acc, f.Placer))
}
