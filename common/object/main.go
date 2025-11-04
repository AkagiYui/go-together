// Package object 提供了一些对象操作的函数
package object

import "reflect"

// HasText 检查对象是否包含文本
func HasText(v any) bool {
	// 如果是指针，获取其指向的值
	if reflect.ValueOf(v).Kind() == reflect.Ptr {
		v = reflect.ValueOf(v).Elem().Interface()
	}

	vValue := reflect.ValueOf(v)

	// 检查是否为字符串类型
	if vValue.Kind() != reflect.String {
		return false
	}

	// 检查是否为空字符串
	return vValue.String() != ""
}
