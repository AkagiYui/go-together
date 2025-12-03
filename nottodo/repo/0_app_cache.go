package repo

import (
	"time"
)

// AppCache 缓存表
// 此模型保留在 models.go 中，因为它被 cache 包使用
type AppCache struct {
	Key       string    `gorm:"column:key;primaryKey;type:varchar(255)" json:"key"`                         // 键
	Value     []byte    `gorm:"column:value;type:jsonb;not null" json:"value"`                              // 值
	ExpiresAt time.Time `gorm:"column:expires_at;type:timestamptz;not null" json:"expiresAt"`               // 过期时间（非空）
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null;default:now()" json:"createdAt"` // 创建时间（非空）
}

// TableName 指定表名
func (AppCache) TableName() string {
	return "app_cache"
}

// SaveCache 保存或更新缓存（UPSERT 操作）
// 如果 key 已存在则更新，否则插入新记录
func SaveCache(cache AppCache) error {
	return DB.Save(&cache).Error
}

// GetCacheByKey 根据 key 查询缓存
// 如果缓存不存在，返回 gorm.ErrRecordNotFound
func GetCacheByKey(key string) (AppCache, error) {
	var cache AppCache
	result := DB.Where("key = ?", key).First(&cache)
	return cache, result.Error
}

// DeleteCache 删除指定 key 的缓存
func DeleteCache(key string) error {
	return DB.Delete(&AppCache{}, "key = ?", key).Error
}

// DeleteExpiredCaches 删除所有过期的缓存项
// 根据 expires_at 字段判断是否过期
func DeleteExpiredCaches() error {
	return DB.Where("expires_at < NOW()").Delete(&AppCache{}).Error
}
