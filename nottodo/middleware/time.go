package middleware

import (
	"fmt"
	"time"

	"github.com/akagiyui/go-together/rest"
)

func TimeConsumeMiddleware() rest.HandlerFunc {
	return func(ctx *rest.Context) {
		beforeTime := time.Now()
		ctx.Next()
		consumeMs := time.Since(beforeTime).Milliseconds()
		ctx.Response.Header("X-Time-Consume", fmt.Sprintf("%dms", consumeMs))
	}
}
