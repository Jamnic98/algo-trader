package engine

type TradeExecutor interface {
	ExecuteTrade(symbol, side string, price float64, qty float64) error
}
