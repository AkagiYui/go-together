// Package middleware 提供HTTP中间件功能
package middleware

import (
	"strings"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/rest"

	"github.com/akagiyui/go-together/arima/config"
	"github.com/akagiyui/go-together/arima/repo"
)

// AuthMiddleware 从请求头中获取 access_key，并验证其有效性
func AuthMiddleware() rest.HandlerFunc {
	return func(ctx *rest.Context) {
		rawToken := ctx.Request.Header.Get("Authorization")
		if accessKey, ok := strings.CutPrefix(rawToken, "Bearer "); ok {
			accessKey = strings.TrimSpace(accessKey)
			if accessKey == "" {
				return
			}

			// 检查是否是管理 API Key
			if accessKey == config.GlobalConfig.ManageAPIKey {
				// 创建一个虚拟的管理员用户
				adminUser := repo.User{
					ID:          0,
					Name:        "Admin",
					IsActive:    true,
					IsSuperuser: true,
				}
				ctx.Set("user", adminUser)
				return
			}

			// 从数据库中查找用户
			user, err := repo.GetUserByAccessKey(accessKey)
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

// RequireSuperuser 用于在需要超级用户权限的路由上使用
func RequireSuperuser() rest.HandlerFunc {
	return func(ctx *rest.Context) {
		userAny, exists := ctx.Get("user")
		if !exists {
			ctx.SetResult(model.Error(model.ErrUnauthorized))
			ctx.Abort()
			return
		}

		user, ok := userAny.(repo.User)
		if !ok || !user.IsSuperuser {
			ctx.SetResult(model.Error(model.ErrUnauthorized, "Not enough permissions"))
			ctx.Abort()
			return
		}
	}
}
