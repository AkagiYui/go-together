// Package system 提供系统设置相关的服务
package system

import (
	"log"
	"time"

	"github.com/akagiyui/go-together/common/task"
	"github.com/akagiyui/go-together/nottodo/cache"
	"github.com/akagiyui/go-together/nottodo/repo"
)

// IsAllowRegistrationCacheKey 是否允许注册的缓存键
const IsAllowRegistrationCacheKey = "setting:is_allow_registration"

// GetIsAllowRegistration 获取是否允许注册的设置
type GetIsAllowRegistration struct{}

// Do 执行获取是否允许注册的业务逻辑
func (r GetIsAllowRegistration) Do() (allowed any, err error) {
	// read from cache
	if err := cache.Get(IsAllowRegistrationCacheKey, &allowed); err == nil {
		return allowed, nil
	}

	defer func() {
		task.Run(func() {
			cache.Set(IsAllowRegistrationCacheKey, allowed, 5*time.Minute)
		})
	}()
	return repo.GetIsAllowRegistration()
}

// SetIsAllowRegistration 设置是否允许注册
type SetIsAllowRegistration struct {
	Allowed bool `json:"allowed"`
}

// Do 执行设置是否允许注册的业务逻辑
func (r SetIsAllowRegistration) Do() (any, error) {
	err := repo.SetIsAllowRegistration(r.Allowed)
	if err != nil {
		return nil, err
	}

	task.Run(func() {
		if err := cache.Delete(IsAllowRegistrationCacheKey); err != nil && err != cache.ErrCacheMiss {
			log.Println("delete cache error:", err)
		}
	})

	return nil, nil
}
