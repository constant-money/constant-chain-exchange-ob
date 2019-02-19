package messages

import "github.com/constant-money/constant-chain-exchange-ob/orderbook"

type OrderMessage struct {
	Type  string          `json:type`
	Order orderbook.Order `json:order`
}
