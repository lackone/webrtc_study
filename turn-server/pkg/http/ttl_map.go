package http

import (
	"sync"
	"time"
)

// Item 带过期时间的数据项
type Item struct {
	Value      any
	ExpireTime time.Time
}

// TTLMap 带 TTL 的并发安全 Map
type TTLMap struct {
	mu sync.RWMutex
	m  map[string]Item
}

// New 创建一个新的 TTLMap
func NewTTLMap() *TTLMap {
	return &TTLMap{
		m: make(map[string]Item),
	}
}

// Set 设置 key-value，ttl 为过期时间（如 time.Second * 30）
func (tm *TTLMap) Set(key string, value any, ttl time.Duration) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.m[key] = Item{
		Value:      value,
		ExpireTime: time.Now().Add(ttl),
	}
}

// Get 获取 key 对应的值，如果不存在或已过期则返回 (nil, false)
func (tm *TTLMap) Get(key string) (value any, ok bool) {
	tm.mu.RLock()
	item, exists := tm.m[key]
	tm.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.ExpireTime) {
		// 懒删除：在读取时清理过期项（可选）
		tm.mu.Lock()
		delete(tm.m, key)
		tm.mu.Unlock()
		return nil, false
	}

	return item.Value, true
}
