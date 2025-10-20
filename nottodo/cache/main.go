package cache

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/akagiyui/go-together/common/task"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/jackc/pgx/v5/pgtype"
)

// ErrCacheMiss 表示在缓存中没有找到对应的项
var ErrCacheMiss = errors.New("cache: key not found")

// Set 向缓存中存入一个值
// key: 缓存键
// value: 需要被缓存的数据结构，它会被序列化为 JSON
// ttl: 缓存的有效期 (Time To Live)，例如 5 * time.Minute
func Set(key string, value interface{}, ttl time.Duration) error {
	// 1. 将 Go struct 序列化为 JSON []byte
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	// 2. 计算过期时间
	expiresAt := time.Now().Add(ttl)

	// 3. 调用 sqlc 生成的方法
	return repo.Db.SetCache(context.Background(), repo.SetCacheParams{
		Key:   key,
		Value: jsonValue, // jsonValue 是 []byte，sqlc 会正确处理它
		ExpiresAt: pgtype.Timestamptz{
			Time:  expiresAt,
			Valid: true,
		},
	})
}

// Get 从缓存中获取一个值
// key: 缓存键
// target: 一个指针，用于接收反序列化后的数据 (例如 &MyStruct{})
func Get(key string, target interface{}) error {
	// 1. 调用 sqlc 生成的方法
	result, err := repo.Db.GetCache(context.Background(), key)
	if err != nil {
		// 如果数据库返回 "no rows"，我们将其转换为我们自定义的 ErrCacheMiss
		if errors.Is(err, sql.ErrNoRows) {
			return ErrCacheMiss
		}
		return err
	}

	// 2. 检查缓存是否已过期
	if time.Now().After(result.ExpiresAt.Time) {
		// 删除已过期的键
		task.Run(func() {
			repo.Db.DeleteCache(context.Background(), key)
		})

		return ErrCacheMiss
	}

	// 3. 将从数据库取出的 JSON []byte 反序列化到 target 指针中
	return json.Unmarshal(result.Value, target)
}

// Delete 从缓存中移除一个键
func Delete(key string) error {
	return repo.Db.DeleteCache(context.Background(), key)
}

// PurgeExpired 会清理所有过期的缓存项。
// 可以在一个后台的 goroutine 中定时运行。
func PurgeExpired() error {
	return repo.Db.PurgeExpiredCache(context.Background())
}
