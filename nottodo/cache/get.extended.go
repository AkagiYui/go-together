package cache

// GetInt64 获取一个 int64 值
func GetInt64(key string) (int64, error) {
	var value int64
	err := Get(key, &value)
	return value, err
}
