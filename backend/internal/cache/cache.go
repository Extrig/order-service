package cache

import (
	"github.com/Extrig/order-service/internal/model"
	"sync"
)

var mu sync.RWMutex
var orders = make(map[string]model.Order)

func Get(orderUID string) (model.Order, bool) {
	mu.RLock()
	defer mu.RUnlock()
	order, ok := orders[orderUID]
	return order, ok
}

func Set(order model.Order) {
	mu.Lock()
	defer mu.Unlock()
	orders[order.OrderUID] = order
}

func SetAll(all map[string]model.Order) {
	mu.Lock()
	defer mu.Unlock()
	orders = all
}
