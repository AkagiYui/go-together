package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// 结构体字段信息缓存
type fieldInfo struct {
	Index int
	Name  string
	Rules []ruleInfo
}

type ruleInfo struct {
	Name  string
	Param string
}

var (
	structCache      = make(map[reflect.Type][]fieldInfo)
	structCacheMutex sync.RWMutex
)

// ValidateStruct 自动校验结构体
// 根据 struct tag 中的 "validate" 标签进行校验
//
// 支持的校验规则：
//   - required: 必填
//   - min=N: 最小值（数值）或最小长度（字符串）
//   - max=N: 最大值（数值）或最大长度（字符串）
//   - len=N: 精确长度（字符串）
//   - email: 邮箱格式
//   - url: URL 格式
//   - alpha: 只包含字母
//   - alphanum: 只包含字母和数字
//   - numeric: 只包含数字
//   - oneof=val1 val2 val3: 值必须是指定值之一
//   - regexp=pattern: 正则表达式匹配
//
// 示例：
//
//	type CreateUserRequest struct {
//	    Username string `json:"username" validate:"required,min=3,max=20,alphanum"`
//	    Email    string `json:"email" validate:"required,email"`
//	    Age      int    `json:"age" validate:"required,min=1,max=150"`
//	}
//
//	func (r *CreateUserRequest) Validate() error {
//	    return validation.ValidateStruct(r)
//	}
func ValidateStruct(v interface{}) error {
	val := reflect.ValueOf(v)

	// 如果是指针，获取其指向的值
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 必须是结构体
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("ValidateStruct 只能用于结构体类型")
	}

	typ := val.Type()

	// 获取或构建字段信息缓存
	fields := getStructFields(typ)

	// 收集所有错误
	var errs []error

	// 遍历所有字段进行校验
	for _, field := range fields {
		fieldValue := val.Field(field.Index)
		fieldName := field.Name

		// 对每个规则进行校验
		for _, rule := range field.Rules {
			err := validateField(fieldValue, rule.Name, rule.Param, fieldName)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	// 使用 errors.Join 合并所有错误
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

// getStructFields 获取结构体字段信息（带缓存）
func getStructFields(typ reflect.Type) []fieldInfo {
	structCacheMutex.RLock()
	if fields, ok := structCache[typ]; ok {
		structCacheMutex.RUnlock()
		return fields
	}
	structCacheMutex.RUnlock()

	// 构建字段信息
	var fields []fieldInfo

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// 跳过未导出的字段
		if !field.IsExported() {
			continue
		}

		// 获取 validate tag
		validateTag := field.Tag.Get("validate")
		if validateTag == "" || validateTag == "-" {
			continue
		}

		// 解析规则
		rules := parseRules(validateTag)
		if len(rules) == 0 {
			continue
		}

		// 获取字段名（优先使用 json tag）
		fieldName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" && parts[0] != "-" {
				fieldName = parts[0]
			}
		}

		fields = append(fields, fieldInfo{
			Index: i,
			Name:  fieldName,
			Rules: rules,
		})
	}

	// 缓存结果
	structCacheMutex.Lock()
	structCache[typ] = fields
	structCacheMutex.Unlock()

	return fields
}

// parseRules 解析校验规则
// 例如: "required,min=3,max=20,alphanum" -> [{required, ""}, {min, "3"}, {max, "20"}, {alphanum, ""}]
func parseRules(tag string) []ruleInfo {
	var rules []ruleInfo

	parts := strings.Split(tag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// 检查是否有参数（如 min=3）
		if idx := strings.Index(part, "="); idx != -1 {
			rules = append(rules, ruleInfo{
				Name:  strings.TrimSpace(part[:idx]),
				Param: strings.TrimSpace(part[idx+1:]),
			})
		} else {
			rules = append(rules, ruleInfo{
				Name:  part,
				Param: "",
			})
		}
	}

	return rules
}

// validateField 校验单个字段
func validateField(field reflect.Value, ruleName string, param string, fieldName string) error {
	// 首先检查是否是自定义校验器
	if customValidator, ok := getValidator(ruleName); ok {
		return customValidator(field, param, fieldName)
	}

	// 内置校验规则
	switch ruleName {
	case "required":
		return validateRequired(field, fieldName)

	case "min":
		return validateMin(field, param, fieldName)

	case "max":
		return validateMax(field, param, fieldName)

	case "len":
		return validateLen(field, param, fieldName)

	case "email":
		return validateEmail(field, fieldName)

	case "url":
		return validateURL(field, fieldName)

	case "alpha":
		return validateAlpha(field, fieldName)

	case "alphanum":
		return validateAlphaNum(field, fieldName)

	case "numeric":
		return validateNumeric(field, fieldName)

	case "oneof":
		return validateOneOf(field, param, fieldName)

	case "regexp":
		return validateRegexp(field, param, fieldName)

	default:
		return fmt.Errorf("未知的校验规则: %s", ruleName)
	}
}

// validateRequired 校验必填
func validateRequired(field reflect.Value, fieldName string) error {
	switch field.Kind() {
	case reflect.String:
		if strings.TrimSpace(field.String()) == "" {
			return fmt.Errorf("%s不能为空", fieldName)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 数值类型的 required 只检查是否为零值
		if field.Int() == 0 {
			return fmt.Errorf("%s不能为空", fieldName)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if field.Uint() == 0 {
			return fmt.Errorf("%s不能为空", fieldName)
		}
	case reflect.Float32, reflect.Float64:
		if field.Float() == 0 {
			return fmt.Errorf("%s不能为空", fieldName)
		}
	case reflect.Bool:
		// bool 类型的 required 检查是否为 false
		if !field.Bool() {
			return fmt.Errorf("%s不能为空", fieldName)
		}
	default:
		// 其他类型检查是否为零值
		if field.IsZero() {
			return fmt.Errorf("%s不能为空", fieldName)
		}
	}
	return nil
}

// validateMin 校验最小值/最小长度
func validateMin(field reflect.Value, param string, fieldName string) error {
	minVal, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("min 参数必须是整数: %s", param)
	}

	switch field.Kind() {
	case reflect.String:
		length := len([]rune(field.String()))
		if length < minVal {
			return fmt.Errorf("%s长度不能少于%d个字符", fieldName, minVal)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() < int64(minVal) {
			return fmt.Errorf("%s不能小于%d", fieldName, minVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if field.Uint() < uint64(minVal) {
			return fmt.Errorf("%s不能小于%d", fieldName, minVal)
		}
	default:
		return fmt.Errorf("min 规则不支持 %s 类型", field.Kind())
	}
	return nil
}

// validateMax 校验最大值/最大长度
func validateMax(field reflect.Value, param string, fieldName string) error {
	maxVal, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("max 参数必须是整数: %s", param)
	}

	switch field.Kind() {
	case reflect.String:
		length := len([]rune(field.String()))
		if length > maxVal {
			return fmt.Errorf("%s长度不能超过%d个字符", fieldName, maxVal)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() > int64(maxVal) {
			return fmt.Errorf("%s不能大于%d", fieldName, maxVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if field.Uint() > uint64(maxVal) {
			return fmt.Errorf("%s不能大于%d", fieldName, maxVal)
		}
	default:
		return fmt.Errorf("max 规则不支持 %s 类型", field.Kind())
	}
	return nil
}

// validateLen 校验精确长度
func validateLen(field reflect.Value, param string, fieldName string) error {
	expectedLen, err := strconv.Atoi(param)
	if err != nil {
		return fmt.Errorf("len 参数必须是整数: %s", param)
	}

	if field.Kind() != reflect.String {
		return fmt.Errorf("len 规则只支持字符串类型")
	}

	length := len([]rune(field.String()))
	if length != expectedLen {
		return fmt.Errorf("%s长度必须为%d个字符", fieldName, expectedLen)
	}
	return nil
}

// validateEmail 校验邮箱格式
func validateEmail(field reflect.Value, fieldName string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("email 规则只支持字符串类型")
	}
	return Email(field.String(), fieldName)
}

// validateURL 校验 URL 格式
func validateURL(field reflect.Value, fieldName string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("url 规则只支持字符串类型")
	}
	return URL(field.String(), fieldName)
}

// validateAlpha 校验只包含字母
func validateAlpha(field reflect.Value, fieldName string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("alpha 规则只支持字符串类型")
	}
	return Alpha(field.String(), fieldName)
}

// validateAlphaNum 校验只包含字母和数字
func validateAlphaNum(field reflect.Value, fieldName string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("alphanum 规则只支持字符串类型")
	}
	return AlphaNumeric(field.String(), fieldName)
}

// validateNumeric 校验只包含数字
func validateNumeric(field reflect.Value, fieldName string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("numeric 规则只支持字符串类型")
	}
	return Numeric(field.String(), fieldName)
}

// validateOneOf 校验值是否在指定集合中
func validateOneOf(field reflect.Value, param string, fieldName string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("oneof 规则只支持字符串类型")
	}

	value := field.String()
	allowed := strings.Fields(param) // 使用空格分隔

	return In(value, allowed, fieldName)
}

// validateRegexp 校验正则表达式匹配
func validateRegexp(field reflect.Value, param string, fieldName string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("regexp 规则只支持字符串类型")
	}
	return Match(field.String(), param, fieldName)
}
