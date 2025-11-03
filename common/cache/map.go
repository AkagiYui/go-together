// Package cache 提供了线程安全的缓存数据结构
package cache

import "sync"

// Map 线程安全的缓存 map，适用于「频繁操作同一批键」的场景
type Map[K comparable, V any] struct {
	data map[K]V
	mu   sync.RWMutex
}

// NewMap 创建新的缓存 map
func NewMap[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		data: make(map[K]V),
	}
}

// Get 获取值，返回值和是否存在
func (c *Map[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.data[key]
	return value, exists
}

// Set 设置值
func (c *Map[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// GetOrSet 获取值，如果不存在则使用 factory 函数创建并设置
func (c *Map[K, V]) GetOrSet(key K, factory func() V) V {
	// 先尝试读取
	c.mu.RLock()
	if value, exists := c.data[key]; exists {
		c.mu.RUnlock()
		return value
	}
	c.mu.RUnlock()

	// 升级为写锁
	c.mu.Lock()
	defer c.mu.Unlock()

	// 双重检查
	if value, exists := c.data[key]; exists {
		return value
	}

	// 创建并设置新值
	value := factory()
	c.data[key] = value
	return value
}

// Len 返回缓存大小
func (c *Map[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// Clear 清空缓存
func (c *Map[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[K]V)
}

// Range 遍历所有键值对，fn 返回 false 时停止遍历
func (c *Map[K, V]) Range(fn func(key K, value V) bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for key, value := range c.data {
		if !fn(key, value) {
			break
		}
	}
}

// Keys 返回所有键的切片
func (c *Map[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]K, 0, len(c.data))
	for key := range c.data {
		keys = append(keys, key)
	}
	return keys
}

// Values 返回所有值的切片
func (c *Map[K, V]) Values() []V {
	c.mu.RLock()
	defer c.mu.RUnlock()

	values := make([]V, 0, len(c.data))
	for _, value := range c.data {
		values = append(values, value)
	}
	return values
}

// ToMap 返回数据的副本
func (c *Map[K, V]) ToMap() map[K]V {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[K]V, len(c.data))
	for key, value := range c.data {
		result[key] = value
	}
	return result
}
