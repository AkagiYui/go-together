package middleware

import (
	"strings"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/nottodo/cache"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

// AuthMiddleware 从请求头中获取 token，并验证其有效性
func AuthMiddleware() rest.HandlerFunc {
	return func(ctx *rest.Context) {
		// 验证 token
		rawToken := ctx.Request.Header.Get("Authorization")
		if token, ok := strings.CutPrefix(rawToken, "Bearer "); ok {
			userId, err := cache.GetInt64("auth_token:" + token)
			if err != nil {
				return
			}
			user, err := repo.GetUserById(userId)
			if err != nil {
				return
			}
			ctx.Set("user", user)
		}
	}
}

// RequireAuth 用于在需要认证的路由上使用，验证请求是否已认证
func RequireAuth() rest.HandlerFunc {
	return func(ctx *rest.Context) {
		if _, exists := ctx.Get("user"); !exists {
			ctx.SetResult(model.Error(model.ErrUnauthorized))
			ctx.Abort()
			return
		}
	}
}
