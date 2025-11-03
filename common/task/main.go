// Package task 提供了安全执行 goroutine 的工具函数
package task

// Options 定义了任务执行的选项
type Options struct {
	OnPanic func(any)
}

// Run 安全地执行 goroutine
func Run(f func()) {
	RunWithOptions(f, nil)
}

// RunWithOptions 使用指定选项安全地执行 goroutine
func RunWithOptions(f func(), opts *Options) {
	go func() {
		if opts == nil {
			opts = &Options{}
		}
		defer func() {
			if err := recover(); err != nil {
				if opts.OnPanic != nil {
					opts.OnPanic(err)
				}
			}
		}()
		f()
	}()
}
