package messages

import "github.com/ninjadotorg/constant-exchange-ob/orderbook"

type OrderMessage struct {
	Type  string          `json:type`
	Order orderbook.Order `json:order`
}
