package engine

type Order struct {
	BotID  string
	Symbol string
	Side   Side
	Qty    float64
}
