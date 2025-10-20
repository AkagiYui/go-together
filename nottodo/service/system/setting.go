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

const IsAllowRegistrationCacheKey = "is_allow_registration"

type GetIsAllowRegistration struct{}

func (r *GetIsAllowRegistration) Handle(ctx *rest.Context) {
	allowed, err := repo.GetIsAllowRegistration()
	if err != nil {
		ctx.SetResult(model.InternalError(err))
		return
	}

	task.Run(func() {
		cache.Set(IsAllowRegistrationCacheKey, allowed, 5*time.Minute)
	})

	ctx.SetResult(model.Success(allowed))
}

type SetIsAllowRegistration struct {
	Allowed bool `json:"allowed"`
}

func (r *SetIsAllowRegistration) Handle(ctx *rest.Context) {
	if err := repo.SetIsAllowRegistration(r.Allowed); err != nil {
		ctx.SetResult(model.InternalError(err))
		return
	}

	task.Run(func() {
		if err := cache.Delete(IsAllowRegistrationCacheKey); err != nil && err != cache.ErrCacheMiss {
			log.Println("delete cache error:", err)
		}
	})

	ctx.SetResult(model.Success(nil))
}
