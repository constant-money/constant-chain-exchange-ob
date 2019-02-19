package orderbook

type Order struct {
	ID     int     `json:"id"`
	Price  float64 `json:"price"`
	Size   float64 `json:"size"`
	Side   string  `json:"side"`
	Symbol string  `json:"symbol"`
	Type   string  `json:"type"`
	Time   int64   `json:"time"`
}
