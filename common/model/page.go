package model

import "reflect"

type PageData struct {
	Total int64 `json:"total"`
	List  any   `json:"list"`
}

func Page(total int64, list any) PageData {
	if reflect.ValueOf(list).IsNil() && total == 0 {
		list = []any{}
	}
	return PageData{
		Total: total,
		List:  list,
	}
}
