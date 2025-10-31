package middleware

import (
	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/rest"
)

func AuthMiddleware() rest.HandlerFunc {
	return func(ctx *rest.Context) {
		// 验证 token
		token := ctx.Request.Header.Get("Authorization")
		if token != "Bearer 123" {
			ctx.SetResult(model.Error(model.ErrUnauthorized, "Unauthorized"))
			ctx.Abort()
			return
		}
	}
}
