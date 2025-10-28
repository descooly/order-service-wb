package cache

import (
	"sync"

	"github.com/descooly/order-service-wb/internal/model"
)

type OrderCache struct {
	Mu    sync.RWMutex
	Cache map[string]model.OrderStruct
}

func New() *OrderCache {
	return &OrderCache{
		Cache: make(map[string]model.OrderStruct),
	}
}

func (o *OrderCache) Set(order model.OrderStruct) {
	o.Mu.Lock()
	defer o.Mu.Unlock()
	o.Cache[order.OrderUID] = order
}

func (o *OrderCache) Get(orderUID string) (model.OrderStruct, bool) {
	o.Mu.RLock()
	defer o.Mu.RUnlock()
	res, ok := o.Cache[orderUID]
	return res, ok
}

func (o *OrderCache) Len() int {
	o.Mu.RLock()
	defer o.Mu.RUnlock()
	return len(o.Cache)
}
