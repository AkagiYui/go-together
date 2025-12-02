package middleware

import (
	"net/http"

	"github.com/akagiyui/go-together/rest"
)

// CorsMiddleware 创建CORS中间件
func CorsMiddleware(allowOrigin string) rest.HandlerFunc {
	return func(ctx *rest.Context) {
		ctx.Response.Header("Access-Control-Allow-Origin", allowOrigin)
		ctx.Response.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		ctx.Response.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		ctx.Response.Header("Access-Control-Allow-Credentials", "true")
		ctx.Response.Header("Access-Control-Max-Age", "86400")

		// 处理预检请求
		if ctx.Method == http.MethodOptions {
			ctx.SetStatusCode(http.StatusNoContent)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
