package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Required 校验字段不能为空
func Required(value string, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s不能为空", fieldName)
	}
	return nil
}

// MinLength 校验字符串最小长度
func MinLength(value string, min int, fieldName string) error {
	// 使用 rune 计数以正确处理中文等多字节字符
	length := len([]rune(value))
	if length < min {
		return fmt.Errorf("%s长度不能少于%d个字符", fieldName, min)
	}
	return nil
}

// MaxLength 校验字符串最大长度
func MaxLength(value string, max int, fieldName string) error {
	// 使用 rune 计数以正确处理中文等多字节字符
	length := len([]rune(value))
	if length > max {
		return fmt.Errorf("%s长度不能超过%d个字符", fieldName, max)
	}
	return nil
}

// LengthRange 校验字符串长度范围
func LengthRange(value string, min int, max int, fieldName string) error {
	length := len([]rune(value))
	if length < min || length > max {
		return fmt.Errorf("%s长度必须在%d-%d个字符之间", fieldName, min, max)
	}
	return nil
}

// Range 校验数值范围
func Range(value int, min int, max int, fieldName string) error {
	if value < min || value > max {
		return fmt.Errorf("%s必须在%d-%d之间", fieldName, min, max)
	}
	return nil
}

// Min 校验数值最小值
func Min(value int, min int, fieldName string) error {
	if value < min {
		return fmt.Errorf("%s不能小于%d", fieldName, min)
	}
	return nil
}

// Max 校验数值最大值
func Max(value int, max int, fieldName string) error {
	if value > max {
		return fmt.Errorf("%s不能大于%d", fieldName, max)
	}
	return nil
}

// Positive 校验数值必须为正数（大于0）
func Positive(value int, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s必须大于0", fieldName)
	}
	return nil
}

// PositiveInt64 校验数值必须为正数（大于0）
func PositiveInt64(value int64, fieldName string) error {
	if value <= 0 {
		return fmt.Errorf("%s必须大于0", fieldName)
	}
	return nil
}

// NonNegative 校验数值必须为非负数（大于等于0）
func NonNegative(value int, fieldName string) error {
	if value < 0 {
		return fmt.Errorf("%s不能为负数", fieldName)
	}
	return nil
}

// Email 校验邮箱格式
// 使用简化的正则表达式，符合大多数常见邮箱格式
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func Email(value string, fieldName string) error {
	if value == "" {
		return nil // 空值不校验，如需校验空值请配合 Required 使用
	}
	if !emailRegex.MatchString(value) {
		return fmt.Errorf("%s格式不正确", fieldName)
	}
	return nil
}

// URL 校验 URL 格式
func URL(value string, fieldName string) error {
	if value == "" {
		return nil // 空值不校验，如需校验空值请配合 Required 使用
	}

	parsedURL, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("%s格式不正确", fieldName)
	}

	// 检查是否有 scheme 和 host
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return fmt.Errorf("%s格式不正确", fieldName)
	}

	return nil
}

// In 校验值是否在指定的集合中
func In(value string, allowed []string, fieldName string) error {
	for _, v := range allowed {
		if value == v {
			return nil
		}
	}
	return fmt.Errorf("%s的值不在允许的范围内", fieldName)
}

// NotIn 校验值是否不在指定的集合中
func NotIn(value string, forbidden []string, fieldName string) error {
	for _, v := range forbidden {
		if value == v {
			return fmt.Errorf("%s的值不允许为%s", fieldName, value)
		}
	}
	return nil
}

// Match 校验字符串是否匹配指定的正则表达式
func Match(value string, pattern string, fieldName string) error {
	if value == "" {
		return nil // 空值不校验
	}

	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		return fmt.Errorf("%s校验失败: 正则表达式错误", fieldName)
	}

	if !matched {
		return fmt.Errorf("%s格式不正确", fieldName)
	}

	return nil
}

// Alpha 校验字符串只包含字母
var alphaRegex = regexp.MustCompile(`^[a-zA-Z]+$`)

func Alpha(value string, fieldName string) error {
	if value == "" {
		return nil
	}
	if !alphaRegex.MatchString(value) {
		return fmt.Errorf("%s只能包含字母", fieldName)
	}
	return nil
}

// AlphaNumeric 校验字符串只包含字母和数字
var alphaNumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func AlphaNumeric(value string, fieldName string) error {
	if value == "" {
		return nil
	}
	if !alphaNumericRegex.MatchString(value) {
		return fmt.Errorf("%s只能包含字母和数字", fieldName)
	}
	return nil
}

// Numeric 校验字符串只包含数字
var numericRegex = regexp.MustCompile(`^[0-9]+$`)

func Numeric(value string, fieldName string) error {
	if value == "" {
		return nil
	}
	if !numericRegex.MatchString(value) {
		return fmt.Errorf("%s只能包含数字", fieldName)
	}
	return nil
}

// Phone 校验中国大陆手机号格式
var phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

func Phone(value string, fieldName string) error {
	if value == "" {
		return nil
	}
	if !phoneRegex.MatchString(value) {
		return fmt.Errorf("%s格式不正确", fieldName)
	}
	return nil
}

// IDCard 校验中国大陆身份证号格式（简化版，只校验格式不校验校验位）
var idCardRegex = regexp.MustCompile(`^[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]$`)

func IDCard(value string, fieldName string) error {
	if value == "" {
		return nil
	}
	if !idCardRegex.MatchString(value) {
		return fmt.Errorf("%s格式不正确", fieldName)
	}
	return nil
}
