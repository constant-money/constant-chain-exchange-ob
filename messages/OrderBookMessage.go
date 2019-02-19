package messages

type OrderBookMessage struct {
	Type string                 `json:type`
	Data map[string]interface{} `json:data`
}
