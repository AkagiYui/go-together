package model

import "reflect"

// PageData 分页数据结构
type PageData struct {
	Total int64 `json:"total"`
	List  any   `json:"list"`
}

// Page 创建分页数据
func Page(total int64, list any) PageData {
	if reflect.ValueOf(list).IsNil() && total == 0 {
		list = []any{}
	}
	return PageData{
		Total: total,
		List:  list,
	}
}
