package cache

import (
    "sync"
    "time"
)

type entry struct {
    v   any
    exp time.Time
}

type TTLCache struct {
    mu   sync.RWMutex
    data map[string]entry
    ttl  time.Duration
}

func NewTTLCache(ttl time.Duration) *TTLCache {
    return &TTLCache{data: make(map[string]entry), ttl: ttl}
}

func (c *TTLCache) Get(key string) (any, bool) {
    c.mu.RLock()
    e, ok := c.data[key]
    c.mu.RUnlock()
    if !ok { return nil, false }
    if time.Now().After(e.exp) {
        c.mu.Lock(); delete(c.data, key); c.mu.Unlock()
        return nil, false
    }
    return e.v, true
}

func (c *TTLCache) Set(key string, v any) {
    c.mu.Lock()
    c.data[key] = entry{v: v, exp: time.Now().Add(c.ttl)}
    c.mu.Unlock()
}
