package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"github.com/constant-money/constant-chain-exchange-ob/config"
	"github.com/constant-money/constant-chain-exchange-ob/messages"
	"github.com/constant-money/constant-chain-exchange-ob/orderbook"
	"github.com/constant-money/constant-chain-exchange-ob/services"
	"github.com/gin-gonic/gin"
	"os"
	"strconv"
)

var conf *config.Config
var orderbooks = map[string]*orderbook.OrderBook{}

const (
	TOPIC_ORDER     = "order"
	TOPIC_ORDERBOOK = "orderbook"
)

func getOrderBook(symbol string) *orderbook.OrderBook {
	if ob, ok := orderbooks[symbol]; ok {
		return ob
	}
	ob := orderbook.NewOrderbook()
	orderbooks[symbol] = ob
	return ob
}

func main() {
	fmt.Println("Start constant exchange")
	conf = config.GetConfig()
	ctx := context.Background()

	r := gin.Default()
	r.GET("/orderbook", func(c *gin.Context) {
		symbol := c.DefaultQuery("symbol", "none")

		if symbol == "none" {
			c.JSON(200, gin.H{
				"data": nil,
			})
		}

		ob := getOrderBook(symbol)

		c.JSON(200, gin.H{
			"data": ob.OrderBook(),
		})
	})
	r.GET("/order", func(c *gin.Context) {
		id := c.DefaultQuery("id", "none")

		if id == "none" {
			c.JSON(200, gin.H{
				"data": nil,
			})
		}
		iId, err := strconv.Atoi(id)
		if err != nil {
			c.JSON(200, gin.H{
				"data": nil,
			})
		}

		for _, ob := range orderbooks {
			order := ob.GetOrder(iId)
			if order != nil {
				c.JSON(200, gin.H{
					"data": ob.OrderBook(),
				})
			}
		}

		c.JSON(200, gin.H{
			"data": nil,
		})
	})
	go r.Run(fmt.Sprintf(":%s", conf.Port))

	ps := services.InitPubSub(conf.GCProjectID, conf.ENV)
	if ps != nil {
		// todo add Logic
		orderTopic := ps.GetOrCreateTopic(TOPIC_ORDER)
		orderbookTopic := ps.GetOrCreateTopic(TOPIC_ORDERBOOK)
		if orderTopic != nil {
			fmt.Println("Has Order topic..")
			publishChange := func(ob *orderbook.OrderBook) {
				msg := map[string]interface{}{
					"type": "change",
					"data": ob.OrderBook(),
				}

				b, _ := json.Marshal(msg)
				_, err := orderbookTopic.Publish(ctx, &pubsub.Message{Data: b}).Get(ctx)

				if err != nil {
					fmt.Println("[OB CHANGE] publish err", err.Error())
				}
			}

			orderSubscribe := ps.GetOrCreateSubscription("orderbook", orderTopic)
			fmt.Println("Start receive..")
			err := orderSubscribe.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
				msg.Ack()
				fmt.Println("add order", string(msg.Data))
				var orderMessage messages.OrderMessage
				if err := json.Unmarshal(msg.Data, &orderMessage); err != nil {
					fmt.Printf("could not decode message data: %#v", msg)
					return
				}

				ob := getOrderBook(orderMessage.Order.Symbol)

				switch orderMessage.Type {
				case "add":
					{
						if matched, matchedOrder := ob.AddOrder(&orderMessage.Order); matched {
							msg := map[string]interface{}{
								"type": "match",
								"data": map[string]interface{}{
									"taker_order_id": orderMessage.Order.ID,
									"maker_order_id": matchedOrder.ID,
									"price":          orderMessage.Order.Price,
									"size":           orderMessage.Order.Size,
								},
							}

							b, _ := json.Marshal(msg)

							_, err := orderbookTopic.Publish(ctx, &pubsub.Message{Data: b}).Get(ctx)

							if err != nil {
								fmt.Println("[OB MATCH] publish err", err.Error())
							}
						}

						// publish ob change
						go publishChange(ob)
					}
				case "update":
					{
						ob := getOrderBook(orderMessage.Order.Symbol)

						if result := ob.UpdateOrder(&orderMessage.Order); !result {
							// publish ob change
							go publishChange(ob)
						}
					}
				case "remove":
					{
						ob := getOrderBook(orderMessage.Order.Symbol)

						if result := ob.RemoveOrder(&orderMessage.Order); !result {
							// publish ob change
							go publishChange(ob)
						}
					}
				default:
					{
						fmt.Println("Unknown type...")
					}
				}
			})

			if err != nil {
				panic(err)
			}
		}
	}
	os.Exit(0)
}
