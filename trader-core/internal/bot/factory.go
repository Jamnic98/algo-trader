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

// NewBot builds a bot with an explicit executor (used when the account is
// chosen per-bot via AccountManager rather than the factory default).
func (f *BotFactory) NewBot(cfg BotConfig, strategy strategies.Strategy, exec engine.ExecutionEngine) (*Bot, error) {
	return f.newBot(cfg, strategy, exec)
}

func (f *BotFactory) NewPaperBot(cfg BotConfig, strategy strategies.Strategy) (*Bot, error) {
	return f.newBot(cfg, strategy, f.Engine())
}

// NewLiveBot builds a bot whose fills are placed on the exchange via placer
// (the broker for cfg.Exchange). Falls back to the factory default Placer when nil.
func (f *BotFactory) NewLiveBot(cfg BotConfig, strategy strategies.Strategy, placer engine.OrderPlacer) (*Bot, error) {
	return f.NewLiveBotFor(cfg, strategy, placer, f.LiveAccount)
}

// NewLiveBotFor is NewLiveBot with an explicit live account (per exchange).
func (f *BotFactory) NewLiveBotFor(cfg BotConfig, strategy strategies.Strategy, placer engine.OrderPlacer, acc *engine.LiveAccount) (*Bot, error) {
	if placer == nil {
		placer = f.Placer
	}
	if acc == nil {
		acc = engine.NewLiveAccount("0", "0.001")
	}
	return f.newBot(cfg, strategy, engine.NewLiveExecutor(acc, placer))
}
