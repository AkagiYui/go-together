package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/nottodo/config"
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

// 在开发模式下启动交互式终端，支持在服务器运行时执行指令
func runInteractiveShell(mode config.Mode) {
	if mode != config.ModeDev {
		return
	}
	go func() {
		reader := bufio.NewScanner(os.Stdin)
		fmt.Println("[DEV] 交互式命令已启用。输入 help 查看帮助。")
		for {
			fmt.Print("> ")
			if !reader.Scan() {
				return
			}
			line := strings.TrimSpace(reader.Text())
			if line == "" {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) == 0 {
				continue
			}
			switch parts[0] {
			case "help", "h":
				fmt.Println("可用命令:")
				fmt.Println("  user create <username> <nickname> <password>  创建用户")
				fmt.Println("  user passwd <username> <newpassword>         修改用户密码")
			case "user":
				if len(parts) < 2 {
					fmt.Println("用法: user <create|passwd> ...")
					continue
				}
				sub := parts[1]
				ctx := context.Background()
				switch sub {
				case "create":
					if len(parts) < 5 {
						fmt.Println("用法: user create <username> <nickname> <password>")
						continue
					}
					u, err := repo.CreateUser(ctx, parts[2], parts[3], parts[4])
					if err != nil {
						fmt.Println("创建失败:", err)
						continue
					}
					fmt.Printf("创建成功: id=%d, username=%s, nickname=%s\n", u.ID, u.Username, u.Nickname)
				case "passwd":
					if len(parts) < 4 {
						fmt.Println("用法: user passwd <username> <newpassword>")
						continue
					}
					if err := repo.UpdateUserPasswordByUsername(ctx, parts[2], parts[3]); err != nil {
						fmt.Println("修改失败:", err)
					} else {
						fmt.Println("修改成功")
					}
				default:
					fmt.Println("未知子命令:", sub)
				}
			default:
				fmt.Println("未知命令，输入 help 获取帮助")
			}
		}
	}()
}

func main() {
	// 读取配置
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// 初始化数据库
	if err := repo.InitDB(cfg.DSN); err != nil {
		panic(err)
	}

	// 开启交互式终端（仅开发模式）
	runInteractiveShell(cfg.Mode)

	s := rest.NewServer()
	s.Debug = cfg.Mode == config.ModeDev

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
