// Package cache 提供缓存操作的扩展功能
package cache

// GetInt64 获取一个 int64 值
func GetInt64(key string) (int64, error) {
	var value int64
	err := Get(key, &value)
	return value, err
}

// GetSlice 获取一个切片
func GetSlice[T any](key string) ([]T, error) {
	var value []T
	err := Get(key, &value)
	return value, err
}
