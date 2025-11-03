package model

import "fmt"

// GeneralResponse 通用响应结构体，使用时不应该直接构造，而是使用 Success、Error、InternalError 等方法
type GeneralResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
	Code    int    `json:"code"`
}

// Success 返回成功响应
func Success(datas ...any) GeneralResponse {
	var data any
	if len(datas) > 1 {
		data = datas
	} else if len(datas) == 1 {
		data = datas[0]
	}
	return GeneralResponse{
		Message: ErrSuccess.Error(),
		Data:    data,
		Code:    businessCodeMap[ErrSuccess],
	}
}

// Error 返回错误响应
func Error(code BusinessCode, messages ...string) GeneralResponse {
	message := code.Error()
	if len(messages) > 0 {
		message = messages[0]
	}
	return GeneralResponse{
		Message: message,
		Data:    nil,
		Code:    businessCodeMap[code],
	}
}

// InternalError 返回服务器内部错误响应
func InternalError(errors ...error) GeneralResponse {
	for _, err := range errors {
		fmt.Printf("err: %v\n", err) // print
	}

	// hide error message
	return Error(ErrInternalError, "Internal Server Error")
}
