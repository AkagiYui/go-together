package model

const (
	SUCCESS = iota + 10000
	INPUT_ERROR
	NOT_FOUND
	UNAUTHORIZED
)

type GeneralResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
	Code    int    `json:"code"`
}

func Success(data any) GeneralResponse {
	return GeneralResponse{
		Message: "",
		Data:    data,
		Code:    SUCCESS,
	}
}

func Error(code int, message string) GeneralResponse {
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
