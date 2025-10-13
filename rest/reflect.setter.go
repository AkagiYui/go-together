package rest

import (
	"fmt"
	"io"
	"mime/multipart"
	"reflect"
	"strconv"
)

// setFileFieldValue 根据字段类型设置文件值
func setFileFieldValue(fieldValue reflect.Value, fileHeader ...*multipart.FileHeader) error {
	if len(fileHeader) == 0 {
		return nil
	}

	// 如果不是切片，或者是指向 []byte 的切片，只使用第一个文件
	if fieldValue.Kind() != reflect.Slice || fieldValue.Type() == reflect.TypeOf([]byte{}) {
		return setFileScalarValue(fieldValue, fileHeader[0])
	}

	// 创建新的切片
	newSlice := reflect.MakeSlice(fieldValue.Type(), len(fileHeader), len(fileHeader))
	// 为每个元素设置值
	for i, fh := range fileHeader {
		elemValue := newSlice.Index(i)
		if err := setFileScalarValue(elemValue, fh); err != nil {
			return err
		}
	}
	// 设置切片
	fieldValue.Set(newSlice)

	return nil
}

// setFileScalarValue 设置单个文件标量值
func setFileScalarValue(fieldValue reflect.Value, fileHeader *multipart.FileHeader) error {
	// 如果是 *multipart.FileHeader，直接设置
	if fieldValue.Type() == reflect.TypeOf(&multipart.FileHeader{}) {
		fieldValue.Set(reflect.ValueOf(fileHeader))
		return nil
	}
	// 如果是 []byte ，读取文件内容并设置
	if fieldValue.Type() == reflect.TypeOf([]byte{}) {
		file, err := fileHeader.Open()
		if err != nil {
			return err
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		fieldValue.Set(reflect.ValueOf(content))
		return nil
	}
	// 如果是 string ，使用 utf-8 解码 []byte 并设置
	if fieldValue.Kind() == reflect.String {
		file, err := fileHeader.Open()
		if err != nil {
			return err
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		fieldValue.SetString(string(content))
		return nil
	}
	return nil
}

// setFieldValue 根据字段类型设置值
func setFieldValue(fieldValue reflect.Value, values ...string) error {
	if len(values) == 0 {
		return nil
	}

	// 对于非切片类型，只使用第一个值
	if fieldValue.Kind() != reflect.Slice {
		return setScalarValue(fieldValue, values[0])
	}

	// 创建新的切片
	newSlice := reflect.MakeSlice(fieldValue.Type(), len(values), len(values))
	// 为每个元素设置值
	for i, value := range values {
		elemValue := newSlice.Index(i)
		if err := setScalarValue(elemValue, value); err != nil {
			return err
		}
	}
	// 设置切片
	fieldValue.Set(newSlice)

	return nil
}

// setScalarValue 设置标量值
func setScalarValue(fieldValue reflect.Value, value string) error {
	switch fieldValue.Kind() {
	case reflect.String: // 字符串
		fieldValue.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64: // 整数
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64: // 无符号整数
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetUint(uintVal)
	case reflect.Float32, reflect.Float64: // 浮点数
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		fieldValue.SetFloat(floatVal)
	case reflect.Bool: // 布尔值
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		fieldValue.SetBool(boolVal)
	default:
		fmt.Printf("Unsupported field type: %s for value: %s\n", fieldValue.Kind(), value)
		return nil
	}
	return nil
}
