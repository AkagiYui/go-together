一个对 Go1.22 ServeMux 进行封装，我自己用的 [REST](https://en.wikipedia.org/wiki/REST) 框架。

支持路径参数、查询参数、请求头、JSON请求体、表单参数、Context值等的自动绑定。

路由语法与 ServeMux 保持一致，本库未重定义路由语法。

## 支持的标签

- `path` - 路径参数
- `query` - 查询参数
- `header` - 请求头
- `json` - JSON 请求体
- `form` - 表单参数
- `context` - Context.Memory 中的值（用于 handler 链向后传递数据）

```golang
package main

import "github.com/akagiyui/go-together/rest"

type HelloRequest struct {
	AccessToken string `header:"access-token"`
	Type        string `path:"type"`
	PathSuffix  string `path:"suffix"`
	Name        string `query:"name"`
	Age         int    `json:"age"`
}

func (r *HelloRequest) Handle(ctx *rest.Context) {
	fmt.Printf("req: %v\n", r)
	ctx.Result("zgm\n")
}

func main() {
	s := rest.NewServer()

	s.GETFunc("/healthz", func(ctx *rest.Context) {
		ctx.Result("OK")
	})
    s.POST("/hello/{type}/zgm/{suffix...}", &HelloRequest{}) // go1.22

	if err := s.Run(":8080"); err != nil {
		panic(err)
	}
}
```

```shell
curl -X POST "http://localhost:8080/hello/123/zgm/456/ddi.txt?name=akagi" \
-H "Access-Token: 123" \
-H "Content-Type: application/json" \
-d '{"age": 18}'
```

## Context 标签使用示例

`context` 标签用于从 Context.Memory 中自动注入值到 handler 结构体字段，常用于 handler 之间传递数据。

```golang
// 认证前置 handler - 设置用户信息到 Context
type AuthPreHandler struct{
	AccessToken string `header:"access-token"`
}

func (r *AuthPreHandler) Handle(ctx *rest.Context) {
    // 验证 access token
    if r.AccessToken != "123" {
        ctx.Status(http.StatusUnauthorized)
        ctx.Result("Unauthorized")
		ctx.Abort()
        return
    }
    // 从数据库加载用户信息
    user := &User{Name: "ddi", ID: 123}
    ctx.Set("current_user", user)
    ctx.Set("user_id", user.ID)
}

// Handler - 自动注入 Context 中的值
type GetProfileRequest struct {
    CurrentUser *User `context:"current_user"` // 自动从 Context 注入
    UserID      int   `context:"user_id"`      // 自动从 Context 注入
}

func (r *GetProfileRequest) Handle(ctx *rest.Context) {
    // r.CurrentUser 和 r.UserID 已经自动注入
    fmt.Printf("User: %v, ID: %d\n", r.CurrentUser, r.UserID)
}

// 使用
s := rest.NewServer()
s.GET("/profile", &AuthPreHandler{}, &GetProfileRequest{})
```

### Context 标签特性

1. **类型安全**：使用反射进行类型检查，类型不匹配时返回错误
2. **支持指针和值类型**：可以注入指针类型（如 `*User`）或值类型（如 `int`、`string`）
3. **nil 值处理**：
   - 如果 Context 中不存在对应的 key，字段保持零值（与其他标签行为一致）
   - 如果 Context 中的值为 nil，指针类型字段会被设置为 nil
4. **混合使用**：可以与其他标签（`path`、`query`、`header`、`json`、`form`）混合使用
