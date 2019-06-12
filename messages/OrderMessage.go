package messages

import "github.com/incognitochain/incognito-chain-exchange-ob/orderbook"

type OrderMessage struct {
	Type  string          `json:type`
	Order orderbook.Order `json:order`
}
