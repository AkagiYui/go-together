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

## ServiceHandlerInterface - 更纯粹的业务处理接口

`ServiceHandlerInterface` 是一个更加纯粹的业务处理接口，它对 HTTP 层透明，只需要实现业务逻辑的核心内容。

### 接口定义

```golang
type ServiceHandlerInterface interface {
    Do() (any, error)
}
```

### 特点

- **Do 函数**只需实现业务逻辑，不需要关心 HTTP 细节
- **第一个返回值**：被视为响应体（response body），会自动设置到 `Context.Result`
- **第二个返回值**：被视为错误（error），会自动设置到 `Context.Status`
- 支持所有标签（`path`、`query`、`header`、`json`、`form`、`context`）
- 支持 `Validator` 接口进行数据校验

### 使用示例

```golang
package main

import (
    "errors"
    "github.com/akagiyui/go-together/rest"
)

// 简单的 ServiceHandler
type HelloServiceHandler struct {
    Name string `json:"name"`
}

func (h HelloServiceHandler) Do() (any, error) {
    if h.Name == "" {
        return nil, errors.New("name is required")
    }
    return map[string]string{
        "message": "Hello, " + h.Name,
    }, nil
}

// 带验证的 ServiceHandler
type CreateUserServiceHandler struct {
    Username string `json:"username"`
    Email    string `json:"email"`
}

func (h CreateUserServiceHandler) Validate() error {
    if h.Username == "" {
        return errors.New("username is required")
    }
    if h.Email == "" {
        return errors.New("email is required")
    }
    return nil
}

func (h CreateUserServiceHandler) Do() (any, error) {
    // 实现业务逻辑
    return map[string]any{
        "id":       123,
        "username": h.Username,
        "email":    h.Email,
    }, nil
}

// 带路径参数的 ServiceHandler
type GetUserServiceHandler struct {
    UserID int64 `path:"id"`
}

func (h GetUserServiceHandler) Do() (any, error) {
    if h.UserID <= 0 {
        return nil, errors.New("invalid user id")
    }
    return map[string]any{
        "id":   h.UserID,
        "name": "User " + fmt.Sprint(h.UserID),
    }, nil
}

func main() {
    server := rest.NewServer()

    // 使用 HandleServ 注册路由
    server.HandleServ("/hello", http.MethodPost, &HelloServiceHandler{})

    // 或使用便捷方法
    server.PostServ("/users", &CreateUserServiceHandler{})
    server.GetServ("/users/{id}", &GetUserServiceHandler{})

    server.Run(":8080")
}
```

### 便捷方法

与 `HandlerInterface` 类似，`ServiceHandlerInterface` 也提供了便捷方法：

- `GetServ(path, handlers...)` - GET 请求
- `PostServ(path, handlers...)` - POST 请求
- `PutServ(path, handlers...)` - PUT 请求
- `DeleteServ(path, handlers...)` - DELETE 请求
- `PatchServ(path, handlers...)` - PATCH 请求
- `AnyServ(path, handlers...)` - 任意 HTTP 方法

### HandlerInterface vs ServiceHandlerInterface

| 特性 | HandlerInterface | ServiceHandlerInterface |
|------|------------------|-------------------------|
| 方法签名 | `Handle(*Context)` | `Do() (any, error)` |
| HTTP 感知 | 需要手动处理 Context | 对 HTTP 层透明 |
| 返回值处理 | 手动调用 `ctx.SetResult()` | 自动设置到 Context |
| 错误处理 | 手动调用 `ctx.SetStatus()` | 自动设置到 Context |
| 适用场景 | 需要精细控制 HTTP 响应 | 纯业务逻辑处理 |
| 参数绑定 | 支持 | 支持 |
| 数据校验 | 支持 Validator | 支持 Validator |
