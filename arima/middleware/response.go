package middleware

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/rest"
)

// ResponseWrapperMiddleware 响应包装中间件
func ResponseWrapperMiddleware() rest.HandlerFunc {
	return func(ctx *rest.Context) {
		ctx.Next()

		// 检测 ctx.Status
		if ctx.Status != nil {
			if ctx.Status != model.ErrSuccess {
				businessCodeObj := ctx.Status.(model.BusinessCode)
				httpStatusCode := model.HTTPStatus(businessCodeObj)

				ctx.SetStatusCode(httpStatusCode)
				if httpStatusCode < 500 {
					ctx.SetResult(model.Error(businessCodeObj))
				} else {
					ctx.SetResult(model.InternalError())
					fmt.Printf("500: %s\n", businessCodeObj.Error())
				}
				return
			}
		}

		// 如果已经是 GeneralResponse,则无需再次封装
		if obj, ok := ctx.Result.(model.GeneralResponse); ok {
			businessCodeObj := model.BusinessCodeFromInt(obj.Code)
			ctx.SetStatusCode(model.HTTPStatus(businessCodeObj))

			if !slices.Contains([]model.BusinessCode{model.ErrSuccess, model.ErrInternalError}, businessCodeObj) {
				fmt.Printf("500: %s\n", obj.Message)
			}
		} else {
			if ctx.StatusCode == http.StatusBadRequest {
				ctx.SetResult(model.Error(model.ErrInputError, "Invalid request"))
			}
		}
	}
}
