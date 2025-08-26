package cache

import (
	"sync"
	"time"
)

type entry[T any] struct {
	val       T
	expiresAt time.Time
}

// TTLCache Simple in-memory TTL cache, concurrency-safe.
// Zero-value is not ready; use New.
type TTLCache[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]entry[V]
	ttl  time.Duration
}

func New[K comparable, V any](ttl time.Duration) *TTLCache[K, V] {
	if ttl <= 0 {
		ttl = 10 * time.Second
	}
	return &TTLCache[K, V]{
		data: make(map[K]entry[V]),
		ttl:  ttl,
	}
}

func (c *TTLCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	e, ok := c.data[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(e.expiresAt) {
		var zero V
		if ok {
			c.mu.Lock()
			delete(c.data, key)
			c.mu.Unlock()
		}
		return zero, false
	}
	return e.val, true
}

func (c *TTLCache[K, V]) Set(key K, val V) {
	e := entry[V]{val: val, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Lock()
	c.data[key] = e
	c.mu.Unlock()
}

// GetOrSet returns existing value if fresh, otherwise determines and sets using supplier.
func (c *TTLCache[K, V]) GetOrSet(key K, supplier func() (V, error)) (V, error) {
	if v, ok := c.Get(key); ok {
		return v, nil
	}
	// compute outside lock
	v, err := supplier()
	if err != nil {
		var zero V
		return zero, err
	}
	c.Set(key, v)
	return v, nil
}
