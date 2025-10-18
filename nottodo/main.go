package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/akagiyui/go-together/common/model"
    "github.com/akagiyui/go-together/nottodo/middleware"
    "github.com/akagiyui/go-together/nottodo/repo"
    "github.com/akagiyui/go-together/nottodo/service"
    "github.com/akagiyui/go-together/rest"
)

func AuthMiddleware() rest.HandlerFunc {
    return func(ctx *rest.Context) {
        // 验证 token
        token := ctx.Request.Header.Get("Authorization")
        if token != "Bearer 123" {
            ctx.SetResult(model.Error(model.UNAUTHORIZED, "Unauthorized"))
            ctx.Abort()
            return
        }
    }
}

func runCLI() bool {
    if len(os.Args) < 2 {
        return false
    }
    if os.Args[1] != "user" {
        return false
    }
    // 初始化数据库（从环境变量读取 DSN）
    if err := repo.InitDB(""); err != nil {
        fmt.Println("初始化数据库失败:", err)
        os.Exit(1)
    }

    userCmd := flag.NewFlagSet("user", flag.ExitOnError)
    if len(os.Args) < 3 {
        fmt.Println("用法: nottodo user <create|passwd> [选项]")
        os.Exit(2)
    }
    sub := os.Args[2]
    switch sub {
    case "create":
        create := flag.NewFlagSet("create", flag.ExitOnError)
        username := create.String("u", "", "用户名")
        nickname := create.String("n", "", "昵称")
        password := create.String("p", "", "密码")
        _ = userCmd // 占位，防止未使用
        _ = create.Parse(os.Args[3:])
        if *username == "" || *password == "" {
            fmt.Println("错误: 用户名与密码不能为空")
            os.Exit(2)
        }
        u, err := repo.CreateUser(*username, *nickname, *password)
        if err != nil {
            fmt.Println("创建用户失败:", err)
            os.Exit(1)
        }
        fmt.Printf("创建成功: id=%d, username=%s, nickname=%s\n", u.ID, u.Username, u.Nickname)
        return true
    case "passwd":
        passwd := flag.NewFlagSet("passwd", flag.ExitOnError)
        username := passwd.String("u", "", "用户名")
        newpass := passwd.String("p", "", "新密码")
        _ = passwd.Parse(os.Args[3:])
        if *username == "" || *newpass == "" {
            fmt.Println("错误: 用户名与新密码不能为空")
            os.Exit(2)
        }
        if err := repo.UpdateUserPasswordByUsername(*username, *newpass); err != nil {
            fmt.Println("修改密码失败:", err)
            os.Exit(1)
        }
        fmt.Println("修改密码成功")
        return true
    default:
        fmt.Println("未知子命令:", sub)
        os.Exit(2)
    }
    return false
}

func main() {
    // CLI 模式
    if runCLI() {
        return
    }

    // 初始化数据库（从环境变量读取 DSN）
    if err := repo.InitDB(""); err != nil {
        panic(err)
    }

    s := rest.NewServer()
    s.Debug = true

    // 设置全局校验错误处理器
    s.SetValidationErrorHandler(func(ctx *rest.Context, err error) {
        ctx.SetResult(model.Error(model.INPUT_ERROR, err.Error()))
    })

    s.UseFunc(middleware.CorsMiddleware(), middleware.TimeConsumeMiddleware())
    s.UseFunc(func(ctx *rest.Context) {
        ctx.Next()
        if obj, ok := ctx.Result.(model.GeneralResponse); ok {
            ctx.Status(model.HttpStatus(obj.Code))
        }
    })

    // 设置 404 处理器
    s.SetNotFoundHandlers(
        func(ctx *rest.Context) {
            ctx.Response.Header("X-Custom", "NotFound")
        },
        func(ctx *rest.Context) {
            ctx.SetResult(model.Error(model.NOT_FOUND, "Route not found"))
        },
    )

    s.GetFunc("/healthz", func(ctx *rest.Context) {})

    v1 := s.Group("/v1")

    todoGroup := v1.Group("/todos", AuthMiddleware())
    {
        todoGroup.Get("", &service.GetTodosRequest{})
        todoGroup.Get("/{id}", &service.GetTodoByIDRequest{})
        todoGroup.Post("", &service.CreateTodoRequest{})
        todoGroup.Put("/{id}", &service.UpdateTodoRequest{})
        todoGroup.Delete("/{id}", &service.DeleteTodoRequest{})
    }

    // 用户管理：仅增删
    userGroup := v1.Group("/users")
    {
        userGroup.Post("", &service.CreateUserRequest{})
        userGroup.Delete("/{id}", &service.DeleteUserRequest{})
    }

    println("🚀 Server starting on http://localhost:8080")
    println("📚 API Documentation:")
    println("  GET    /v1/todos        - 获取所有Todo")
    println("  GET    /v1/todos/{id}   - 获取指定ID的Todo")
    println("  POST   /v1/todos        - 创建Todo")
    println("  PUT    /v1/todos/{id}   - 更新指定ID的Todo")
    println("  DELETE /v1/todos/{id}   - 删除指定ID的Todo")
    println("  POST   /v1/users        - 创建用户")
    println("  DELETE /v1/users/{id}   - 删除用户")

    if err := s.Run(":8080"); err != nil {
        panic(err)
    }
}
