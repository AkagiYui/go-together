package task

import (
	"errors"
)

// CollectError 执行一个函数，如果函数返回错误，则将其加入到dest中。
func CollectError(f func() error, dest *error) {
	if err := f(); err != nil {
		if *dest == nil {
			*dest = err
			return
		}
		*dest = errors.Join(*dest, err)
	}
}
