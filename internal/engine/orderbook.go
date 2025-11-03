package engine

import (
	"sync"
	"time"

	"github.com/emirpasic/gods/trees/redblacktree"
)

type OrderSide string

const (
	Buy  OrderSide = "BUY"
	Sell OrderSide = "SELL"
)

type OrderType string

const (
	Limit  OrderType = "LIMIT"
	Market OrderType = "MARKET"
)

type OrderStatus string

const (
	Pending   OrderStatus = "PENDING"
	Partial   OrderStatus = "PARTIAL"
	Filled    OrderStatus = "FILLED"
	Cancelled OrderStatus = "CANCELLED"
)

// Order represents a trading order
type Order struct {
	ID          string
	UserID      string
	Symbol      string
	Side        OrderSide
	Type        OrderType
	Price       float64
	Quantity    float64
	FilledQty   float64
	Status      OrderStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	SequenceNum int64
}

// PriceLevel represents orders at a specific price point
type PriceLevel struct {
	Price  float64
	Orders []*Order
	Volume float64
}

// OrderBook manages buy and sell orders using red-black trees
type OrderBook struct {
	Symbol   string
	BuyTree  *redblacktree.Tree // Max heap for buys (descending price)
	SellTree *redblacktree.Tree // Min heap for sells (ascending price)
	OrderMap map[string]*Order  // Quick lookup by order ID
	mu       sync.RWMutex
	sequence int64
}

// NewOrderBook creates a new order book for a trading pair
func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol: symbol,
		// Buy orders: highest price first (reverse comparator)
		BuyTree: redblacktree.NewWith(func(a, b interface{}) int {
			priceA := a.(float64)
			priceB := b.(float64)
			if priceA > priceB {
				return -1
			} else if priceA < priceB {
				return 1
			}
			return 0
		}),
		// Sell orders: lowest price first (normal comparator)
		SellTree: redblacktree.NewWith(func(a, b interface{}) int {
			priceA := a.(float64)
			priceB := b.(float64)
			if priceA < priceB {
				return -1
			} else if priceA > priceB {
				return 1
			}
			return 0
		}),
		OrderMap: make(map[string]*Order),
	}
}

// AddOrder adds an order to the order book
func (ob *OrderBook) AddOrder(order *Order) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	ob.sequence++
	order.SequenceNum = ob.sequence

	ob.OrderMap[order.ID] = order

	tree := ob.getTree(order.Side)

	// Get or create price level
	value, found := tree.Get(order.Price)
	var level *PriceLevel

	if found {
		level = value.(*PriceLevel)
	} else {
		level = &PriceLevel{
			Price:  order.Price,
			Orders: make([]*Order, 0),
			Volume: 0,
		}
		tree.Put(order.Price, level)
	}

	level.Orders = append(level.Orders, order)
	level.Volume += order.Quantity
}

// RemoveOrder removes an order from the order book
func (ob *OrderBook) RemoveOrder(orderID string) *Order {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	order, exists := ob.OrderMap[orderID]
	if !exists {
		return nil
	}

	delete(ob.OrderMap, orderID)

	tree := ob.getTree(order.Side)
	value, found := tree.Get(order.Price)
	if !found {
		return order
	}

	level := value.(*PriceLevel)
	remainingQty := order.Quantity - order.FilledQty

	// Remove order from level
	for i, o := range level.Orders {
		if o.ID == orderID {
			level.Orders = append(level.Orders[:i], level.Orders[i+1:]...)
			level.Volume -= remainingQty
			break
		}
	}

	// Remove price level if empty
	if len(level.Orders) == 0 {
		tree.Remove(order.Price)
	}

	return order
}

// GetBestBid returns the highest buy price
func (ob *OrderBook) GetBestBid() (float64, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	if ob.BuyTree.Size() == 0 {
		return 0, false
	}

	node := ob.BuyTree.Left() // Highest price in buy tree
	return node.Key.(float64), true
}

// GetBestAsk returns the lowest sell price
func (ob *OrderBook) GetBestAsk() (float64, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	if ob.SellTree.Size() == 0 {
		return 0, false
	}

	node := ob.SellTree.Left() // Lowest price in sell tree
	return node.Key.(float64), true
}

// GetSpread returns the bid-ask spread
func (ob *OrderBook) GetSpread() float64 {
	bid, hasBid := ob.GetBestBid()
	ask, hasAsk := ob.GetBestAsk()

	if !hasBid || !hasAsk {
		return 0
	}

	return ask - bid
}

// GetDepth returns the order book depth (top N levels)
func (ob *OrderBook) GetDepth(levels int) (bids, asks []PriceLevel) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	bids = make([]PriceLevel, 0, levels)
	asks = make([]PriceLevel, 0, levels)

	// Get top buy levels
	it := ob.BuyTree.Iterator()
	count := 0
	for it.Next() && count < levels {
		level := it.Value().(*PriceLevel)
		bids = append(bids, *level)
		count++
	}

	// Get top sell levels
	it = ob.SellTree.Iterator()
	count = 0
	for it.Next() && count < levels {
		level := it.Value().(*PriceLevel)
		asks = append(asks, *level)
		count++
	}

	return bids, asks
}

func (ob *OrderBook) getTree(side OrderSide) *redblacktree.Tree {
	if side == Buy {
		return ob.BuyTree
	}
	return ob.SellTree
}

// GetOrder retrieves an order by ID
func (ob *OrderBook) GetOrder(orderID string) (*Order, bool) {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	order, exists := ob.OrderMap[orderID]
	return order, exists
}

// Size returns the total number of orders
func (ob *OrderBook) Size() int {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	return len(ob.OrderMap)
}
