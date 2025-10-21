# common

这是一个零依赖库，用于给我的程序提供一些通用的工具函数和数据结构。

## EnumRegistry

枚举注册器，用于注册和管理枚举值。支持所有可比较类型。

```golang
package main

import "github.com/akagiyui/go-together/common/enum"

type Color int

var colors = enum.NewEnumRegistry[Color]()

var (
	Red   = colors.Register(1)
	Blue  = colors.Register(2)
	Green = colors.Register(3)
)

func main() {
	println(colors.String())        // [0, 1, 2]
	println(colors.Contains(Red))   // true
	println(colors.Contains(3))     // false
}
```
