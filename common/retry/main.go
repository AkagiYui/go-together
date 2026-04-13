package retry

import (
	"errors"
	"time"
)

// FixedDelay 固定时间间隔重试，如果f返回FatalError则不再重试，达到重试上限则返回原始错误，成功返回nil。
func FixedDelay(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; i < attempts; i++ {
		// 立即执行
		err = f()
		// 如果成功,则返回
		if err == nil {
			return nil
		}
		// 如果是FatalError,则不再重试
		if fatalErr, ok := errors.AsType[*FatalError](err); ok {
			return fatalErr.Err
		}
		// 睡眠
		time.Sleep(sleep)
	}
	return err
}
