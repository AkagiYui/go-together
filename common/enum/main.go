// Package enum 提供了枚举类型的注册和管理功能
package enum

import "fmt"

// Registry 枚举注册器 - 支持所有可比较类型
type Registry[T comparable] struct {
	values []T
}

// NewRegistry 创建新的枚举注册器
func NewRegistry[T comparable]() *Registry[T] {
	return &Registry[T]{
		values: make([]T, 0),
	}
}

// Register 注册枚举值
func (r *Registry[T]) Register(value T) T {
	r.values = append(r.values, value)
	return value
}

// Values 获取所有枚举值
func (r *Registry[T]) Values() []T {
	return r.values
}

// Contains 检查是否包含指定值
func (r *Registry[T]) Contains(value T) bool {
	for _, v := range r.values {
		if v == value {
			return true
		}
	}
	return false
}

// String 返回所有枚举值的字符串表示
func (r *Registry[T]) String() string {
	if len(r.values) == 0 {
		return "[]"
	}

	result := "["
	for i, v := range r.values {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%v", v)
	}
	result += "]"
	return result
}

// Len 返回枚举值数量
func (r *Registry[T]) Len() int {
	return len(r.values)
}
