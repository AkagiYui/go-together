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
		afterTime := time.Now()
		consumeMs := afterTime.Sub(beforeTime).Milliseconds()
		fmt.Printf("consume: %dms\n", consumeMs)
		ctx.Response.Header("X-Time-Consume", fmt.Sprintf("%dms", consumeMs))
	}
}
