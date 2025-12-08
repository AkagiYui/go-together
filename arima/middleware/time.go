package middleware

import (
	"log/slog"
	"time"

	"github.com/akagiyui/go-together/rest"
)

// TimeConsumeMiddleware 记录请求耗时的中间件
func TimeConsumeMiddleware() rest.HandlerFunc {
	return func(ctx *rest.Context) {
		start := time.Now()
		ctx.Next()
		duration := time.Since(start)
		slog.Debug("Request completed",
			slog.String("method", ctx.Method),
			slog.String("path", ctx.Endpoint),
			slog.Duration("duration", duration),
		)
	}
}
