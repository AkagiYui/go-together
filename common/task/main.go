package task

type Options struct {
	OnPanic func(any)
}

// Run 安全地执行 goroutine
func Run(f func()) {
	RunWithOptions(f, nil)
}

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
