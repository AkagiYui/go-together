package model

const (
	INPUT_ERROR = 10001
)

type GeneralResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Code    int    `json:"code,omitempty"`
}

func Success(data any) GeneralResponse {
	return GeneralResponse{
		Message: "success",
		Data:    data,
	}
}

func Error(code int, message string) GeneralResponse {
	return GeneralResponse{
		Message: message,
		Code:    code,
	}
}

type PageData struct {
	Total int `json:"total"`
	List  any `json:"list"`
}
