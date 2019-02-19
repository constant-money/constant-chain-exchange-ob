package orderbook

import (
	"fmt"
	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/shopspring/decimal"
	"sort"
	"sync"
)

const SIDE_BUY = "buy"
const SIDE_SELL = "sell"

type OrderBook struct {
	buy                 *redblacktree.Tree
	sell                *redblacktree.Tree
	mutex               sync.Mutex
	orders              map[int]*Order
	onBeforeAddOrder    func(ob *OrderBook, o *Order)
	onBeforeUpdateOrder func(ob *OrderBook, o *Order)
	onBeforeRemoveOrder func(ob *OrderBook, o *Order)
	onAfterAddOrder     func(ob *OrderBook, o *Order)
	onAfterUpdateOrder  func(ob *OrderBook, o *Order)
	onAfterRemoveOrder  func(ob *OrderBook, o *Order)
}

func NewOrderbook() *OrderBook {
	ob := &OrderBook{
		buy:    redblacktree.NewWith(Float64ComparatorAsc),
		sell:   redblacktree.NewWith(Float64ComparatorDesc),
		orders: make(map[int]*Order),
	}
	return ob
}

func (ob *OrderBook) getTrees(side string) (*redblacktree.Tree, *redblacktree.Tree) {
	var tree *redblacktree.Tree
	var reverseTree *redblacktree.Tree

	if side == SIDE_BUY {
		tree = ob.buy
		reverseTree = ob.sell
	} else {
		tree = ob.sell
		reverseTree = ob.buy
	}

	return tree, reverseTree
}

func (ob *OrderBook) findMatching(tree *redblacktree.Tree, price float64, size float64) *Order {
	if value, ok := tree.Get(price); ok {
		orders := value.([]*Order)
		sort.Slice(orders[:], func(i, j int) bool {
			return orders[i].Time < orders[j].Time
		})

		var isMatching = false
		var matchOrder *Order
		// find matching order
		for _, order := range orders {
			if order.Size == size && order.Price == price {
				// todo matching
				isMatching = true
				matchOrder = order
				break
			}
		}
		if isMatching {
			return matchOrder
		}
		return nil
	}
	return nil
}

func (ob *OrderBook) removeOrder(tree *redblacktree.Tree, order *Order) bool {
	if value, ok := tree.Get(order.Price); ok {
		orders := value.([]*Order)

		var matchedIdx = -1
		for idx, o := range orders {
			if o.ID == order.ID {
				// todo matching
				matchedIdx = idx
				break
			}
		}

		if matchedIdx != -1 {
			if len(orders) == 1 {
				tree.Remove(order.Price)
			} else {
				newOrders := append(orders[:matchedIdx], orders[matchedIdx+1:]...)
				tree.Put(order.Price, newOrders)
			}

			return true
		}

		return false
	}
	return false
}

func (ob *OrderBook) addOrder(tree *redblacktree.Tree, order *Order) bool {
	if value, ok := tree.Get(order.Price); ok {
		orders := value.([]*Order)
		newOrders := append(orders, order)
		tree.Put(order.Price, newOrders)
	} else {
		orders := make([]*Order, 0)
		orders = append(orders, order)
		tree.Put(order.Price, orders)
	}
	return true
}

func (ob *OrderBook) updateOrder(tree *redblacktree.Tree, order *Order) bool {
	var status bool = true
	if oldOrder, ok := ob.orders[order.ID]; ok {
		if oldOrder.Price != order.Price {
			status = ob.removeOrder(tree, order)
		}
	}
	if status {
		status = ob.addOrder(tree, order)
	}
	return status
}

func (ob *OrderBook) AddOrder(o *Order) (bool, *Order) {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()
	if ob.onBeforeAddOrder != nil {
		ob.onBeforeAddOrder(ob, o)
	}

	var status bool = false
	var order *Order = nil

	tree, reverseTree := ob.getTrees(o.Side)

	matchedOrder := ob.findMatching(reverseTree, o.Price, o.Size)

	if matchedOrder != nil {
		// matching case
		fmt.Println("call remove")
		if ob.removeOrder(reverseTree, matchedOrder) {
			status = true
			order = matchedOrder
		} else {
			panic("[OB removeOrder] fail...")
		}
	} else {
		// un-match case
		if ob.addOrder(tree, o) {
			// add order to mapping
			ob.orders[o.ID] = o
			return false, nil
		} else {
			panic("[OB addOrder] fail...")
		}
	}

	if ob.onAfterAddOrder != nil {
		ob.onAfterAddOrder(ob, o)
	}

	return status, order
}

func (ob *OrderBook) UpdateOrder(o *Order) bool {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()
	var status bool = false

	if ob.onBeforeUpdateOrder != nil {
		ob.onBeforeUpdateOrder(ob, o)
	}
	//
	tree, _ := ob.getTrees(o.Side)
	status = ob.updateOrder(tree, o)

	if ob.onAfterUpdateOrder != nil {
		ob.onAfterUpdateOrder(ob, o)
	}
	return status
}

func (ob *OrderBook) RemoveOrder(o *Order) bool {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()
	var status bool = false

	if ob.onBeforeRemoveOrder != nil {
		ob.onBeforeRemoveOrder(ob, o)
	}
	//
	tree, _ := ob.getTrees(o.Side)
	status = ob.removeOrder(tree, o)

	if ob.onAfterRemoveOrder != nil {
		ob.onAfterRemoveOrder(ob, o)
	}
	return status
}

func (ob *OrderBook) TotalBuyOrders() int {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	size := 0
	values := ob.buy.Values()
	for _, value := range values {
		size += len(value.([]*Order))
	}
	return size
}

func (ob *OrderBook) TotalSellOrders() int {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	size := 0
	values := ob.sell.Values()
	for _, value := range values {
		size += len(value.([]*Order))
	}
	return size
}

func (ob *OrderBook) OrderBook() map[string]interface{} {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	buy := make([]interface{}, 0)
	sell := make([]interface{}, 0)

	buyIt := ob.buy.Iterator()
	for buyIt.Next() {
		orders := buyIt.Value().([]*Order)
		sum := decimal.NewFromFloat(0)
		for _, o := range orders {
			sum = sum.Add(decimal.NewFromFloat(o.Size))
		}
		buy = append(buy, []interface{}{fmt.Sprintf("%g", buyIt.Key().(float64)), sum.String(), buyIt.Value()})
	}

	sellIt := ob.sell.Iterator()
	for sellIt.Next() {
		orders := sellIt.Value().([]*Order)
		sum := decimal.NewFromFloat(0)
		for _, o := range orders {
			sum = sum.Add(decimal.NewFromFloat(o.Size))
		}
		sell = append(sell, []interface{}{fmt.Sprintf("%g", sellIt.Key().(float64)), sum.String(), sellIt.Value()})
	}

	return map[string]interface{}{
		"buy":  buy,
		"sell": sell,
	}
}

func (ob *OrderBook) GetOrder(id int) *Order {
	ob.mutex.Lock()
	defer ob.mutex.Unlock()

	order, ok := ob.orders[id]
	if ok {
		return order
	}

	return nil
}
