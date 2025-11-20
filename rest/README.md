# REST

一个基于 Go 1.22 ServeMux 的轻量级 RESTful API 框架，支持自动参数绑定、中间件、路由组等功能。

> [!CAUTION]
> 该库目前仍在积极开发中，API 随时会发生不兼容变更，请谨慎使用于生产环境。

如果你习惯通过示例学习，[快速上手](#快速上手) 和 [示例代码](#示例代码) 部分会帮助你快速了解如何使用该库。

## 目录

- [目录](#目录)
- [安装](#安装)
- [快速上手](#快速上手)
- [功能详解](#功能详解)
  - [参数绑定](#参数绑定)
    - [支持的标签类型](#支持的标签类型)
    - [完整参数绑定示例](#完整参数绑定示例)
  - [如何接收请求](#如何接收请求)
    - [HandlerInterface](#handlerinterface)
    - [ServiceHandlerInterface](#servicehandlerinterface)
  - [中间件](#中间件)
    - [函数式中间件](#函数式中间件)
    - [结构体中间件](#结构体中间件)
    - [单独某个路由使用中间件](#单独某个路由使用中间件)
  - [路由组](#路由组)
    - [路由组中间件](#路由组中间件)
  - [Context 对象](#context-对象)
    - [请求信息](#请求信息)
    - [响应设置](#响应设置)
    - [Context 内存存储](#context-内存存储)
  - [错误处理](#错误处理)
  - [数据验证](#数据验证)
- [调试模式](#调试模式)
- [示例代码](#示例代码)
  - [上传文件](#上传文件)
  - [ServiceHandlerInterface 结合中间件](#servicehandlerinterface-结合中间件)

## 安装

```shell
go get github.com/akagiyui/go-together/rest@latest
```

## 快速上手

以下是一个最简单的 REST 应用示例：

```go
package main

import (
    "fmt"
    "github.com/akagiyui/go-together/rest"
)

// 定义请求处理器
type HelloHandler struct {
    Name string `query:"name"`
    Age  int    `json:"age"`
}

// 实现 Handle 方法
func (h HelloHandler) Handle(ctx *rest.Context) {
    message := fmt.Sprintf("Hello %s, you are %d years old!", h.Name, h.Age)
    ctx.SetResult(message)
}

func main() {
    // 创建服务器
    server := rest.NewServer()

    // 注册路由
    server.Post("/hello", &HelloHandler{})

    // 启动服务器
    if err := server.Run(":8080"); err != nil {
        panic(err)
    }
}
```

测试请求：
```shell
curl -X POST "http://localhost:8080/hello?name=Alice" \
  -H "Content-Type: application/json" \
  -d '{"age": 25}'
```

## 功能详解

### 参数绑定

REST 支持通过结构体标签自动绑定各种类型的请求参数：

#### 支持的标签类型

- `path` - 路径参数
- `query` - 查询参数
- `header` - 请求头
- `json` - JSON 请求体
- `form` - 表单参数
- `context` - Context.Memory 中的值

#### 完整参数绑定示例

```go
type CompleteHandler struct {
    // 路径参数
    UserID   int64  `path:"id"`
    Category string `path:"category"`

    // 查询参数
    Page  int    `query:"page"`
    Limit int    `query:"limit"`

    // 请求头
    Token     string `header:"Authorization"`
    UserAgent string `header:"User-Agent"`

    // JSON 请求体
    Name  string `json:"name"`
    Email string `json:"email"`

    // Context 值（用于中间件传递数据）
    CurrentUser User `context:"user"`
}

func (h CompleteHandler) Handle(ctx *rest.Context) {
    // 所有参数已自动绑定到结构体字段
    ctx.SetResult(map[string]interface{}{
        "user_id":      h.UserID,
        "category":     h.Category,
        "page":         h.Page,
        "limit":        h.Limit,
        "token":        h.Token,
        "user_agent":   h.UserAgent,
        "name":         h.Name,
        "email":        h.Email,
        "current_user": h.CurrentUser,
    })
}

func main() {
    server := rest.NewServer()
    server.Put("/users/{id}/categories/{category}", &CompleteHandler{})
    server.Run(":8080")
}
```

`context` 标签的详细用法请参考 [Context 内存存储](#context-内存存储) 部分。

### 如何接收请求

你需要为结构体实现 `HandlerInterface` 接口或 `ServiceHandlerInterface` 接口，
为他们提供处理请求的能力。

#### HandlerInterface

标准的处理器接口，提供完整的 HTTP 上下文控制：

```go
type HandlerInterface interface {
    Handle(*Context)
}
```

**使用示例：**

```go
type UserHandler struct {
    ID int64 `path:"id"`
}

func (h UserHandler) Handle(ctx *rest.Context) {
    user := getUserByID(h.ID)
    if user == nil {
        ctx.SetStatusCode(404)
        ctx.SetResult("User not found")
        return
    }
    ctx.SetResult(user)
}

// 注册路由
server.Get("/users/{id}", &UserHandler{})
```

REST 提供了部分 HTTP 方法的便捷注册方式：

```go
server.Get("/users", &GetUsersHandler{}) // server.Handle("GET", "/users", &GetUsersHandler{})
server.Post("/users", &CreateUserHandler{}) // server.Handle("POST", "/users", &CreateUserHandler{})
server.Put("/users/{id}", &UpdateUserHandler{}) // server.Handle("PUT", "/users/{id}", &UpdateUserHandler{})
server.Delete("/users/{id}", &DeleteUserHandler{}) // server.Handle("DELETE", "/users/{id}", &DeleteUserHandler{})
server.Patch("/users/{id}", &PatchUserHandler{}) // server.Handle("PATCH", "/users/{id}", &PatchUserHandler{})
```

如果没有你需要的 HTTP 方法，可以使用通用的 `Handle` 注册：

```go
server.Handle("YOUR_METHOD", "/users", &OptionsHandler{})
```

如果你需要处理简单的请求，可以使用函数式处理器(就像 gin 一样)，但你需要手动处理请求参数：

```go
server.GetFunc("/health", func(ctx *rest.Context) {
    ctx.SetResult("OK")
})
```

#### ServiceHandlerInterface

除了 `HandlerInterface` ，REST 还提供了一个更简洁的服务处理器接口，
`ServiceHandlerInterface` 接口更适合纯业务逻辑处理，无需关心 HTTP 细节。

这或许可以为你节省一些 controller 层的代码：

```go
type ServiceHandlerInterface interface {
    Do() (any, error)
}
```

**特点：**
- **Do 函数**只需实现业务逻辑，不需要关心 HTTP 细节
- **第一个返回值**：被视为响应体，会自动设置到 `Context.Result`
- **第二个返回值**：被视为错误，会自动设置到 `Context.Status`
- 支持与 `HandlerInterface` 一致的所有标签（`path`、`query`、`header`、`json`、`form`、`context`）

得益于 Golang 的隐式接口实现，你的业务代码可以**完全不包含** REST 相关依赖。

**使用示例：**

```go
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

    // 使用 GetServ 而不是 Get
    server.GetServ("/users/{id}", &GetUserServiceHandler{})

    server.Run(":8080")
}
```

> [!NOTE]
> 由于 Golang 的无协变特性，你实现的Do方法需要返回 `any` 类型，不能直接返回具体类型（如 `User`），否则会导致编译错误。
> 
> 你必须时刻注意你返回的值是否是你期望的内容，避免暴露过多不希望暴露的内部细节。


同样地，REST 也提供了部分 HTTP 方法的便捷注册方式：

```go
server.GetServ("/users", &GetUsersServiceHandler{})
server.PostServ("/users", &CreateUserServiceHandler{})
server.PutServ("/users/{id}", &UpdateUserServiceHandler{})
server.DeleteServ("/users/{id}", &DeleteUserServiceHandler{})
server.PatchServ("/users/{id}", &PatchUserServiceHandler{})
```

### 中间件

REST 支持中间件功能，
你可以在中间件中处理请求和响应，或进行认证、日志记录等操作。

#### 函数式中间件

中间件可以是一个函数，与处理器函数类似，
但你可以使用 `ctx.Next()` 方法将控制权传递给下一个中间件或处理器，
或者使用 `ctx.Abort()` 方法中止请求处理。

```go
func main() {
    server := rest.NewServer()

    // 添加全局中间件
    server.UseFunc(func(ctx *rest.Context) {
        fmt.Println("Before request")
        ctx.Next() // 继续执行后续处理器
        fmt.Println("After request")
    })

    // 添加全局 CORS 中间件
    server.UseFunc(func(ctx *rest.Context) {
        ctx.Response.Header("Access-Control-Allow-Origin", "*")
        ctx.Next()
    })

    server.Post("/hello", &HelloHandler{})
    server.Run(":8080")
}
```

#### 结构体中间件

你也可以使用结构体形式的中间件，需要实现 `HandlerInterface` 接口。
同样地，你可以使用结构体标签绑定请求参数。

```go
type AuthMiddleware struct {
    Token string `header:"Authorization"`
}

func (m AuthMiddleware) Handle(ctx *rest.Context) {
    if m.Token == "" {
        ctx.SetStatusCode(401)
        ctx.SetResult("Unauthorized")
        ctx.Abort() // 中止后续处理器执行
        return
    }

    // 验证 token 并设置用户信息
    user := validateToken(m.Token)
    ctx.Set("user", user)
    ctx.Next() // 继续执行，如果 ctx.Next() 是中间件的最后一行语句，你可以省略它，比如在这里你完全可以删除这行代码
}

func main() {
    server := rest.NewServer()
    server.Use(&AuthMiddleware{}) // 全局认证中间件
    server.Post("/protected", &ProtectedHandler{})
    server.Run(":8080")
}
```

#### 单独某个路由使用中间件

你可以单独为某个路由添加中间件，只需在注册路由时传入中间件结构体：

```go
server.Get("/users", &AuthMiddleware{}, &GetUsersHandler{})
server.Post("/users", &AuthMiddleware{}, &CreateUserHandler{})
```

### 路由组

路由组允许你组织相关的路由，为部分路由添加前缀，并应用特定的中间件。

你可以使用 `Group` 方法创建路由组，`server` 或其他路由组都可以创建子路由组。

路由组可以嵌套创建，子路由组会继承父路由组的前缀和中间件。

```go
func main() {
    server := rest.NewServer()

    // 创建 API v1 路由组
    v1 := server.Group("/api/v1")
    {
        // 用户相关路由
        userGroup := v1.Group("/users")
        {
            userGroup.Get("", &GetUsersHandler{})
            userGroup.Post("", &CreateUserHandler{})
            userGroup.Get("/{id}", &GetUserHandler{})
            userGroup.Put("/{id}", &UpdateUserHandler{})
            userGroup.Delete("/{id}", &DeleteUserHandler{})
        }

        // 需要认证的路由组
        authGroup := v1.Group("/auth", &AuthMiddleware{})
        {
            authGroup.Get("/profile", &GetProfileHandler{})
            authGroup.Post("/logout", &LogoutHandler{})
        }
    }

    server.Run(":8080")
}
```

> [!NOTE]
> 以上的 `{}` 只是为了代码块的可读性，并不是必须的语法。

#### 路由组中间件

```go
func main() {
    server := rest.NewServer()

    // 创建带中间件的路由组
    apiGroup := server.Group("/api")
    apiGroup.UseFunc(func(ctx *rest.Context) {
        fmt.Println("API middleware")
        ctx.Next()
    })

    // 子路由组继承父组的中间件
    v1Group := apiGroup.Group("/v1")
    v1Group.UseFunc(func(ctx *rest.Context) {
        fmt.Println("V1 middleware")
        ctx.Next()
    })

    v1Group.Get("/users", &GetUsersHandler{})
    server.Run(":8080")
}
```

### Context 对象

Context 对象提供了丰富的请求和响应处理功能：

#### 请求信息

> [!TIP]
> 你应该优先使用结构体标签绑定请求参数，而不是直接操作 Context.Request 对象。

```go
func (h MyHandler) Handle(ctx *rest.Context) {
    // 请求基本信息
    method := ctx.Request.Method
    path := ctx.Request.Endpoint
    uri := ctx.Request.URI
    host := ctx.Request.Host
    remoteAddr := ctx.Request.RemoteAddr

    // 请求头
    userAgent := ctx.Request.Header.Get("User-Agent")

    // 查询参数
    page := ctx.Request.Query.Get("page")

    // 路径参数
    userID := ctx.Request.PathParams["id"]

    // 请求体
    body := ctx.FillBody() // 获取原始请求体
}
```

> [!NOTE]
> `ctx.FillBody()` 方法会读取并缓存请求体，后续调用不会重复读取，你可以安全地多次调用该方法。

#### 响应设置

```go
func (h MyHandler) Handle(ctx *rest.Context) {
    // 设置响应状态码
    ctx.SetStatusCode(200)

    // 设置响应头
    ctx.Response.Header("Content-Type", "application/json")
    ctx.Response.Header("X-Custom-Header", "value")

    // 设置响应体
    ctx.SetResult(map[string]string{
        "message": "success",
    })

    // 设置状态（用于错误处理）
    ctx.SetStatus(errors.New("something went wrong"))
}
```

#### Context 内存存储

Context 提供了线程安全的内存存储，用于在中间件和处理器之间传递数据。

`context` 标签用于将 Context.Memory 中的值自动注入值到处理器结构体字段。

```go
// 认证中间件设置用户信息
type AuthMiddleware struct {
    Token string `header:"Authorization"`
}

func (m AuthMiddleware) Handle(ctx *rest.Context) {
    user := validateToken(m.Token)
    ctx.Set("current_user", user)
    ctx.Set("user_id", user.ID)
    ctx.Next()
}

// 业务处理器自动注入用户信息
type GetProfileHandler struct {
    CurrentUser User `context:"current_user"`
    UserID      int64 `context:"user_id"`
}

func (h GetProfileHandler) Handle(ctx *rest.Context) {
    // h.CurrentUser 和 h.UserID 已自动注入
    ctx.SetResult(map[string]any{
        "user": h.CurrentUser,
        "id":   h.UserID,
    })
}
```

### 错误处理

```go
func main() {
    server := rest.NewServer()

    // 设置全局校验错误处理器
    server.SetValidationErrorHandler(func(ctx *rest.Context, err error) {
        ctx.SetStatusCode(400)
        ctx.SetResult(map[string]string{
            "error": "Validation failed: " + err.Error(),
        })
    })

    // 设置 404 处理器
    server.SetNotFoundHandlers(func(ctx *rest.Context) {
        ctx.SetStatusCode(404)
        ctx.SetResult(map[string]string{
            "error": "Not found",
        })
    })

    server.Run(":8080")
}
```

### 数据验证

框架支持通过 `Validator` 接口进行数据校验：

```go
type CreateUserHandler struct {
    Username string `json:"username"`
    Email    string `json:"email"`
    Age      int    `json:"age"`
}

// 实现 Validator 接口，该方法会在 Handle 方法调用前被自动执行
// 如果返回错误，Handle 方法不会被调用，该方法应当仅处理校验逻辑
// 不应该包含业务逻辑
func (h CreateUserHandler) Validate() error {
    if h.Username == "" {
        return errors.New("username is required")
    }
    if len(h.Username) < 3 {
        return errors.New("username must be at least 3 characters")
    }
    if h.Email == "" {
        return errors.New("email is required")
    }
    if !strings.Contains(h.Email, "@") {
        return errors.New("invalid email format")
    }
    if h.Age < 0 || h.Age > 150 {
        return errors.New("age must be between 0 and 150")
    }
    return nil
}

func (h CreateUserHandler) Handle(ctx *rest.Context) {
    ctx.SetResult(map[string]string{
        "message": "User created successfully",
    })
}
```

## 调试模式

启用调试模式可以查看所有注册的路由：

```go
func main() {
    server := rest.NewServer()
    server.Debug = true // 启用调试模式

    server.Get("/users", &GetUsersHandler{})
    server.Post("/users", &CreateUserHandler{})

    server.Run(":8080")
    // 输出：
    // [    GET] /users                           --> main.GetUsersHandler (1 handlers)
    // [   POST] /users                           --> main.CreateUserHandler (1 handlers)
}
```

## 示例代码

### 上传文件

你可以使用 `multipart/form-data` 来上传文件，REST 会自动将文件内容绑定到结构体字段。

你可以使用 `[]byte`、`string` 或 `*multipart.FileHeader` 类型来接收上传的文件内容。
使用 `[]byte` 时，REST 会将文件内容读取到内存中。
使用 `string` 时，REST 会将文件内容读取为 UTF-8 字符串，便于直接处理文本文件。

> [!WARNING]
> 如果你使用 `[]byte` 和 `string` 类型来接收文件内容，REST 会将整个文件读入内存。
> 
> 大文件可能会导致内存占用过高，建议使用 `*multipart.FileHeader` 类型来处理大文件上传。

```go
type UploadHandler struct {
    File1 []byte                `form:"file1"`
    File2 string                `form:"file2"`
    File3 *multipart.FileHeader `form:"file3"`
}

func (h *UploadHandler) Handle(ctx *rest.Context) {
    // 处理上传的文件
}
```

### ServiceHandlerInterface 结合中间件

该方式可使你的业务代码完全不依赖 REST 框架，
你可以更轻松地在网络请求、定时任务和命令行，甚至不同项目间复用业务代码。

```go
// main.go
package main

// 响应包装中间件
type ResponseWrapperMiddleware struct{}
func (m ResponseWrapperMiddleware) Handle(ctx *rest.Context) {
    ctx.Next()
    originalResult := ctx.Result
    ctx.SetResult(map[string]any{
        "status": "success",
        "data":   originalResult,
    })
}

func main() {
    server := rest.NewServer()
    server.Use(&ResponseWrapperMiddleware{})
    server.GetServ("/time", &GetTimeServiceHandler{})
    server.Run(":8080")
}
```

```go
// service.go
package main

type GetTimeServiceHandler struct{}
func (h GetTimeServiceHandler) Do() (any, error) {
    return map[string]string{
        "time": time.Now().Format(time.RFC3339),
    }, nil
}
```
