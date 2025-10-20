package validation

// Validate 校验结构体，根据其字段上的 `validate` 标签进行校验。
// 如果校验通过，返回 nil；否则返回包含所有错误信息的 error。
//
// 这是 ValidateStruct 的别名，便于更直观地调用：
//   err := validation.Validate(req)
func Validate(v interface{}) error {
	return ValidateStruct(v)
}
