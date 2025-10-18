package validation

import (
	"reflect"
	"sync"
)

// ValidatorFunc 自定义校验函数类型
// field: 字段的反射值
// param: tag 中的参数（如 "min=3" 中的 "3"）
// fieldName: 字段名称（用于错误信息）
type ValidatorFunc func(field reflect.Value, param string, fieldName string) error

// 全局校验器注册表
var (
	customValidators = make(map[string]ValidatorFunc)
	validatorMutex   sync.RWMutex
)

// RegisterValidator 注册自定义校验规则
// name: 校验规则名称（在 tag 中使用）
// fn: 校验函数
//
// 示例：
//
//	validation.RegisterValidator("password", func(field reflect.Value, param string, fieldName string) error {
//	    password := field.String()
//	    if len(password) < 8 {
//	        return fmt.Errorf("%s长度不能少于8个字符", fieldName)
//	    }
//	    // 检查是否包含大写字母
//	    if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
//	        return fmt.Errorf("%s必须包含大写字母", fieldName)
//	    }
//	    return nil
//	})
func RegisterValidator(name string, fn ValidatorFunc) {
	validatorMutex.Lock()
	defer validatorMutex.Unlock()
	customValidators[name] = fn
}

// getValidator 获取自定义校验器（内部使用）
func getValidator(name string) (ValidatorFunc, bool) {
	validatorMutex.RLock()
	defer validatorMutex.RUnlock()
	fn, ok := customValidators[name]
	return fn, ok
}

// UnregisterValidator 取消注册自定义校验规则（主要用于测试）
func UnregisterValidator(name string) {
	validatorMutex.Lock()
	defer validatorMutex.Unlock()
	delete(customValidators, name)
}

// ClearValidators 清空所有自定义校验规则（主要用于测试）
func ClearValidators() {
	validatorMutex.Lock()
	defer validatorMutex.Unlock()
	customValidators = make(map[string]ValidatorFunc)
}
