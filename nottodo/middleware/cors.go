package middleware

import (
	"net/http"

	"github.com/akagiyui/go-together/rest"
)

// CorsMiddleware 跨域资源共享中间件
func CorsMiddleware(allowOrigin string) rest.HandlerFunc {
	return func(ctx *rest.Context) {
		ctx.Response.Header("Access-Control-Allow-Origin", allowOrigin)
		ctx.Response.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Response.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		ctx.Response.Header("Access-Control-Allow-Credentials", "true")
		ctx.Response.Header("Access-Control-Max-Age", "86400")

		if ctx.Request.Method == "OPTIONS" {
			ctx.SetStatusCode(http.StatusNoContent)
			ctx.SetResult(nil)
			ctx.Abort()
			return
		}
	}
}
