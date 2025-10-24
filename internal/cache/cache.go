package my_cache

import (
	"project/internal"
	"sync"
)

type OrderCache struct {
	Mu    sync.RWMutex
	Cache map[string]internal.OrderStruct
}

func New() *OrderCache {
	return &OrderCache{
		Cache: make(map[string]internal.OrderStruct),
	}
}

func (o *OrderCache) Set(order internal.OrderStruct) {
	o.Mu.Lock()
	defer o.Mu.Unlock()
	o.Cache[order.OrderUID] = order
}

func (o *OrderCache) Get(orderUID string) (internal.OrderStruct, bool) {
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
