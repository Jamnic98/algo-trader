package engine

type ExecutionEngine interface {
	ExecuteTrade(Order) (*Fill, error)
}
