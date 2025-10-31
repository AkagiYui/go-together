// Package model 提供了一些常见的业务异常，工作区内项目应使用该包下的异常类型
package model

import (
	"errors"
	"net/http"
)

// BusinessCode 业务状态码
type BusinessCode error

var (
	ErrSuccess       BusinessCode = errors.New("success")        // 成功
	ErrInputError    BusinessCode = errors.New("input error")    // 输入错误(参数校验失败等)
	ErrNotFound      BusinessCode = errors.New("not found")      // 未找到
	ErrUnauthorized  BusinessCode = errors.New("unauthorized")   // 未授权
	ErrInternalError BusinessCode = errors.New("internal error") // 服务器内部错误
)

var businessCodeMap = map[BusinessCode]int{
	ErrSuccess:       0,
	ErrInputError:    1,
	ErrNotFound:      2,
	ErrUnauthorized:  3,
	ErrInternalError: 4,
}
var businessCodeReverseMap = map[int]BusinessCode{}

func init() {
	for code, value := range businessCodeMap {
		businessCodeReverseMap[value] = code
	}
}

func BusinessCodeFromInt(code int) BusinessCode {
	if businessCode, ok := businessCodeReverseMap[code]; ok {
		return businessCode
	}
	return ErrInternalError
}

// statusMap BusinessCode->HTTP状态码 映射表
var statusMap = map[BusinessCode]int{
	ErrSuccess:       http.StatusOK,
	ErrInputError:    http.StatusBadRequest,
	ErrNotFound:      http.StatusNotFound,
	ErrUnauthorized:  http.StatusUnauthorized,
	ErrInternalError: http.StatusInternalServerError,
}

// HttpStatus 将业务错误码转换为 HTTP 状态码
func HttpStatus(code BusinessCode) int {
	if status, ok := statusMap[code]; ok {
		return status
	}
	return http.StatusInternalServerError
}
