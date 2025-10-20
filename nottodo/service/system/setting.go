package system

import (
	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

type GetIsAllowRegistration struct{}

func (r *GetIsAllowRegistration) Handle(ctx *rest.Context) {
	println("GetIsAllowRegistration")
	allowed, err := repo.GetIsAllowRegistration()
	if err != nil {
		ctx.SetResult(model.InternalError(err))
		return
	}
	ctx.SetResult(model.Success(allowed))
}

type SetIsAllowRegistration struct {
	Allowed bool `json:"allowed"`
}

func (r *SetIsAllowRegistration) Handle(ctx *rest.Context) {
	println("SetIsAllowRegistration")
	if err := repo.SetIsAllowRegistration(r.Allowed); err != nil {
		ctx.SetResult(model.InternalError(err))
		return
	}
	ctx.SetResult(model.Success(nil))
}
