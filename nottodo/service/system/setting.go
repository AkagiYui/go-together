package system

import (
	"log"
	"time"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/common/task"
	"github.com/akagiyui/go-together/nottodo/cache"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

const IsAllowRegistrationCacheKey = "setting:is_allow_registration"

type GetIsAllowRegistration struct{}

func (r GetIsAllowRegistration) Handle(ctx *rest.Context) {
	allowed, err := r.Do()
	if err != nil {
		ctx.SetResult(model.InternalError(err))
		return
	}
	ctx.SetResult(model.Success(allowed))
}

func (r GetIsAllowRegistration) Do() (allowed bool, err error) {
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

type SetIsAllowRegistration struct {
	Allowed bool `json:"allowed"`
}

func (r SetIsAllowRegistration) Handle(ctx *rest.Context) {
	if err := r.Do(); err != nil {
		ctx.SetResult(model.InternalError(err))
		return
	}
	ctx.SetResult(model.Success(nil))
}

func (r SetIsAllowRegistration) Do() error {
	err := repo.SetIsAllowRegistration(r.Allowed)
	if err != nil {
		return err
	}

	task.Run(func() {
		if err := cache.Delete(IsAllowRegistrationCacheKey); err != nil && err != cache.ErrCacheMiss {
			log.Println("delete cache error:", err)
		}
	})

	return nil
}
