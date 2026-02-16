package engine

type TradeExecutor interface {
	ExecuteTrade(botID, symbol string, side Side, price, qty float64) (*Fill, error)
}

type ExecutionEngine interface {
	PlaceOrder(order Order) error
}
