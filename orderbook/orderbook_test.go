package orderbook

import (
	"encoding/json"
	"fmt"
	"github.com/ninjadotorg/constant-exchange-ob/utils"
	"sync"
	"testing"
	"time"
)

var ai = 0
var mutex sync.Mutex

func randomOrder(side string, price float64, size float64) *Order {
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

	return &Order{
		ID:    ai,
		Price: p,
		Size:  s,
		Side:  sd,
		Type:  "limit",
		Time:  time.Now().Unix(),
	}
}

func dump(ob *OrderBook) {
	obJson, _ := json.Marshal(ob.OrderBook())
	fmt.Println(string(obJson))
}

func TestAddOrder(t *testing.T) {
	ob := NewOrderbook()
	o := randomOrder("buy", 1, 5)

	ob.AddOrder(o)
	if actualValue := ob.buy.Size(); actualValue != 1 {
		t.Errorf("Got %v expected %v", actualValue, 1)
	}

	//dump(ob)
}

func TestSingleMatching(t *testing.T) {
	ob := NewOrderbook()

	o := randomOrder("buy", 1, 5)
	ob.AddOrder(o)

	o1 := randomOrder("sell", 1, 5)
	matched, matchedOrder := ob.AddOrder(o1)

	if !matched {
		t.Errorf("Matched got %v expect %v ", matched, true)
	}

	if matched && matchedOrder.ID != o.ID {
		t.Errorf("Matched Order ID got %v expect  %v ", matchedOrder.ID, o.ID)
	}

	if actualValue := ob.buy.Size(); actualValue != 0 {
		t.Errorf("Got %v expected %v", actualValue, 0)
	}
}

func TestMultiMatching(t *testing.T) {
	ob := NewOrderbook()

	// add 6 buy order, 4 fixed, 2 random
	b1 := randomOrder("buy", 1, 5)
	ob.AddOrder(b1)

	b2 := randomOrder("buy", 2, 5)
	ob.AddOrder(b2)

	b3 := randomOrder("buy", 3, 5)
	ob.AddOrder(b3)

	b4 := randomOrder("buy", 4, 5)
	ob.AddOrder(b4)

	b5 := randomOrder("buy", 0, 5)
	ob.AddOrder(b5)

	b6 := randomOrder("buy", 0, 5)
	ob.AddOrder(b6)

	// add 2 sell order with 2 fixed match b2, b4
	s2 := randomOrder("sell", 2, 5)
	matchedS2, matchedOrderS2 := ob.AddOrder(s2)

	s4 := randomOrder("sell", 4, 5)
	matchedS4, matchedOrderS4 := ob.AddOrder(s4)

	if !matchedS2 {
		t.Errorf("Matched S2 got %v expect %v ", matchedS2, true)
	}

	if matchedS2 && matchedOrderS2.ID != b2.ID {
		t.Errorf("Matched Order ID got %v expect  %v ", matchedOrderS2.ID, b2.ID)
	}

	if !matchedS4 {
		t.Errorf("Matched S2 got %v expect %v ", matchedS2, true)
	}

	if matchedS4 && matchedOrderS4.ID != b4.ID {
		t.Errorf("Matched Order ID got %v expect  %v ", matchedOrderS4.ID, b4.ID)
	}

	if actualValue := ob.buy.Size(); actualValue != 4 {
		t.Errorf("Got %v expected %v", actualValue, 4)
	}
}

func TestRandomMatching(t *testing.T) {
	ob := NewOrderbook()

	randomBuy := utils.RandomInt(20, 50)
	randomSell := utils.RandomInt(30, 100)

	matchedCount := 0

	for i := 0; i < randomBuy; i++ {
		ob.AddOrder(randomOrder("buy", 0, 0))
	}

	//dump(ob)
	//fmt.Println("OB buy size", ob.buy.Size())

	for i := 0; i < randomSell; i++ {
		matched, _ := ob.AddOrder(randomOrder("sell", 0, 0))
		if matched {
			matchedCount++
		}
	}

	//dump(ob)
	//fmt.Println("Matched", matchedCount, randomBuy, randomSell)

	if actualValue := ob.TotalBuyOrders(); actualValue != randomBuy-matchedCount {
		t.Errorf("Got %v expected %v, count %v", actualValue, randomBuy-matchedCount, matchedCount)
	}
}

func TestRBT(t *testing.T) {
	ob := NewOrderbook()

	// add 6 buy order, 4 fixed, 2 random
	b1 := randomOrder("buy", 6, 5)
	ob.AddOrder(b1)

	b2 := randomOrder("buy", 12, 5)
	ob.AddOrder(b2)

	b3 := randomOrder("buy", 3, 5)
	ob.AddOrder(b3)

	b4 := randomOrder("buy", 22, 5)
	ob.AddOrder(b4)

	b5 := randomOrder("buy", 8, 5)
	ob.AddOrder(b5)

	b6 := randomOrder("buy", 1, 5)
	ob.AddOrder(b6)

	//dump(ob)

	for idx, key := range ob.buy.Keys() {
		if idx == 0 && key != float64(1) {
			t.Errorf("Element %d got %v expected %v", idx+1, key, 1)
		}
		if idx == 1 && key != float64(3) {
			t.Errorf("Element %d got %v expected %v", idx+1, key, 3)
		}
		if idx == 2 && key != float64(6) {
			t.Errorf("Element %d got %v expected %v", idx+1, key, 6)
		}
		if idx == 3 && key != float64(8) {
			t.Errorf("Element %d got %v expected %v", idx+1, key, 8)
		}
		if idx == 4 && key != float64(12) {
			t.Errorf("Element %d got %v expected %v", idx+1, key, 12)
		}
		if idx == 5 && key != float64(22) {
			t.Errorf("Element %d got %v expected %v", idx+1, key, 22)
		}
	}
}
