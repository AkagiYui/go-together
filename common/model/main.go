package model

import (
	"fmt"
	"net/http"
	"reflect"
)

type BusinessCode int

const (
	SUCCESS BusinessCode = iota + 10000
	INPUT_ERROR
	NOT_FOUND
	UNAUTHORIZED
	INTERNAL_ERROR
)

var statusMap = map[BusinessCode]int{
	SUCCESS:        http.StatusOK,
	INPUT_ERROR:    http.StatusBadRequest,
	NOT_FOUND:      http.StatusNotFound,
	UNAUTHORIZED:   http.StatusUnauthorized,
	INTERNAL_ERROR: http.StatusInternalServerError,
}

// 将业务错误码转换为 HTTP 状态码
func HttpStatus(code BusinessCode) int {
	if status, ok := statusMap[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

type GeneralResponse struct {
	Message string       `json:"message"`
	Data    any          `json:"data"`
	Code    BusinessCode `json:"code"`
}

func Success(data any) GeneralResponse {
	return GeneralResponse{
		Message: "",
		Data:    data,
		Code:    SUCCESS,
	}
}

func Error(code BusinessCode, message string) GeneralResponse {
	return GeneralResponse{
		Message: message,
		Data:    nil,
		Code:    code,
	}
}

func InternalError(errors ...error) GeneralResponse {
	for _, err := range errors {
		fmt.Printf("err: %v\n", err) // print
	}

	// hide error message
	return Error(INTERNAL_ERROR, "Internal Server Error")
}

type PageData struct {
	Total int64 `json:"total"`
	List  any   `json:"list"`
}

func Page(total int64, list any) PageData {
	if IsNil(list) && total == 0 {
		list = []any{}
	}
	return PageData{
		Total: total,
		List:  list,
	}
}

func IsNil(v any) bool {
	return reflect.ValueOf(v).IsNil()
}
