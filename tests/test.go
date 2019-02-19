package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ninjadotorg/constant-exchange-ob/config"
	"github.com/ninjadotorg/constant-exchange-ob/messages"
	"github.com/ninjadotorg/constant-exchange-ob/orderbook"
	"github.com/ninjadotorg/constant-exchange-ob/services"
	"github.com/ninjadotorg/constant-exchange-ob/utils"
	"os"
	"sync"
	"time"
)

const (
	TOPIC_ORDER     = "order"
	TOPIC_ORDERBOOK = "orderbook"
)

var ai = 0
var mutex sync.Mutex

func randomOrder(symbol string, side string, price float64, size float64) *orderbook.Order {
	mutex.Lock()
	p := price
	s := size
	sd := side

	if p == 0 {
		p = utils.RandomFloat64(1, 5)
	}

	if s == 0 {
		s = utils.RandomFloat64(1, 5)
	}

	if sd == "" {
		randomSide := utils.RandomInt(0, 2)
		if randomSide == 1 {
			sd = "sell"
		} else {
			sd = "buy"
		}
	}

	ai++
	mutex.Unlock()

	return &orderbook.Order{
		ID:     ai,
		Price:  p,
		Size:   s,
		Side:   sd,
		Symbol: symbol,
		Type:   "limit",
		Time:   time.Now().Unix(),
	}
}

func main() {
	conf := config.GetConfig()

	ctx := context.Background()

	ps := services.InitPubSub(conf.GCProjectID, conf.ENV)

	orderTopic := ps.GetOrCreateTopic(TOPIC_ORDER)
	orderbookTopic := ps.GetOrCreateTopic(TOPIC_ORDERBOOK)

	var orderbookSubscribe *pubsub.Subscription
	changeSubscribeName := "order"

	symbols := []string{"BTCUSD", "ETHUSD", "BONDUSD"}

	if orderbookTopic != nil {
		orderbookSubscribe = ps.GetOrCreateSubscription(changeSubscribeName, orderbookTopic)
		go orderbookSubscribe.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			msg.Ack()

			var orderbookMessage messages.OrderBookMessage
			if err := json.Unmarshal(msg.Data, &orderbookMessage); err != nil {
				fmt.Printf("could not decode message data: %#v", msg)
				return
			}
			//if orderbookMessage.Type == "match" {
			//	fmt.Println("ob match", string(msg.Data))
			//}
			fmt.Println("ob", string(msg.Data))
		})
	}

	for _, symbol := range symbols {
		totalOrder := utils.RandomInt(1000, 5000)
		for i := 0; i < totalOrder; i++ {
			o := randomOrder(symbol, "", 0, 0)
			msg := map[string]interface{}{
				"type":  "add",
				"order": o,
			}
			msgJson, _ := json.Marshal(msg)
			uuid, err := orderTopic.Publish(ctx, &pubsub.Message{Data: msgJson}).Get(ctx)
			if err != nil {
				fmt.Println("Publish order error", err.Error())
			} else {
				fmt.Println("Add order", string(msgJson), uuid)
			}

			//time.Sleep(10 * time.Millisecond)
		}
	}

	time.Sleep(30 * time.Second)

	err := orderbookSubscribe.Delete(ctx)
	if err != nil {
		fmt.Println("Delete subscription error: ", err.Error())
	}

	os.Exit(0)
}
