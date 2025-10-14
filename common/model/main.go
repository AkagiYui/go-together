package model

import "net/http"

type BusinessCode int

const (
	SUCCESS BusinessCode = iota + 10000
	INPUT_ERROR
	NOT_FOUND
	UNAUTHORIZED
)

var statusMap = map[BusinessCode]int{
	SUCCESS:      http.StatusOK,
	INPUT_ERROR:  http.StatusBadRequest,
	NOT_FOUND:    http.StatusNotFound,
	UNAUTHORIZED: http.StatusUnauthorized,
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

type PageData struct {
	Total int `json:"total"`
	List  any `json:"list"`
}
