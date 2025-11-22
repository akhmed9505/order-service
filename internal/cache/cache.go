package cache

import (
	"errors"
	"sync"

	"github.com/hashicorp/golang-lru"

	"order-service-wbtech/internal/model"
)

type OrderCache struct {
	lru *lru.Cache
	mu  sync.RWMutex
}

const defaultCacheSize = 1000

// New creates a new LRU cache with a default size
func New() (*OrderCache, error) {
	l, err := lru.New(defaultCacheSize)
	if err != nil {
		return nil, errors.New("failed to create LRU cache")
	}
	return &OrderCache{
		lru: l,
	}, nil
}

func (c *OrderCache) Get(orderUID string) (*model.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, ok := c.lru.Get(orderUID)
	if !ok {
		return nil, false
	}

	order, ok := value.(*model.Order)
	return order, ok
}

func (c *OrderCache) Set(order *model.Order) {
	if order == nil || order.OrderUID == "" {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.lru.Add(order.OrderUID, order)
}
