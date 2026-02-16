package engine

type Order struct {
	BotID  string
	Symbol string
	Side   Side
	Price  float64
	Qty    float64
}
