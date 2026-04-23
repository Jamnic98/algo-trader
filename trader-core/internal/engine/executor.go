package engine

type ExecutionEngine interface {
	ExecuteFill(Order) (*Fill, error)
}
