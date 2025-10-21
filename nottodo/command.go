package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"

    "github.com/akagiyui/go-together/common/model"
    "github.com/akagiyui/go-together/nottodo/config"
    usersvc "github.com/akagiyui/go-together/nottodo/service/user"
    "github.com/akagiyui/go-together/rest"
)

// 在开发模式下启动交互式终端，支持在服务器运行时执行指令
// 当前支持命令：noop, adduser
const help = `可用命令:
noop                           占位命令，输出 ok
adduser <username> <password>  新建用户`

func runInteractiveShell(mode config.Mode) {
    if mode != config.ModeDev {
        return
    }
    go func() {
        reader := bufio.NewScanner(os.Stdin)
        fmt.Println("[DEV] 交互式命令已启用。")
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
            cmd := parts[0]
            switch cmd {
            case "noop":
                handleNoop()
            case "adduser":
                handleAddUser(parts[1:])
            default:
                fmt.Println(help)
            }
        }
    }()
}

func handleNoop() {
    fmt.Println("ok")
}

func handleAddUser(args []string) {
    if len(args) != 2 {
        fmt.Println("用法: adduser <username> <password>")
        return
    }
    username := args[0]
    password := args[1]

    req := usersvc.CreateUserRequest{Username: username, Password: password}
    if err := req.Validate(); err != nil {
        fmt.Println("错误: ", err)
        return
    }
    ctx := rest.NewEmptyContext()
    req.Handle(&ctx)
    if resp, ok := ctx.Result.(model.GeneralResponse); ok {
        if resp.Code != model.SUCCESS {
            fmt.Println("错误: ", resp.Message)
            return
        }
        if u, ok := resp.Data.(usersvc.UserResponse); ok {
            fmt.Printf("ok, 用户已创建，ID=%d, 用户名=%s\n", u.ID, username)
        } else {
            fmt.Println("ok")
        }
    } else {
        fmt.Println("ok")
    }
}
