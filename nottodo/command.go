package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"

    "github.com/akagiyui/go-together/nottodo/config"
    "github.com/akagiyui/go-together/nottodo/repo"
    usersvc "github.com/akagiyui/go-together/nottodo/service/user"
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
                fmt.Println("ok")
            case "adduser":
                if len(parts) != 3 {
                    fmt.Println("用法: adduser <username> <password>")
                    continue
                }
                username := parts[1]
                password := parts[2]
                hashed, err := usersvc.HashPassword(password)
                if err != nil {
                    fmt.Println("错误: 密码哈希失败 -", err)
                    continue
                }
                newUser, err := repo.CreateUser(repo.User{Username: username, Password: hashed})
                if err != nil {
                    fmt.Println("错误: 创建用户失败 -", err)
                    continue
                }
                fmt.Printf("ok, 用户已创建，ID=%d, 用户名=%s\n", newUser.ID, newUser.Username)
            default:
                fmt.Println(help)
            }
        }
    }()
}
