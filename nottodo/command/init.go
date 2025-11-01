package command

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/akagiyui/go-together/nottodo/config"
)

// 在开发模式下启动交互式终端，支持在服务器运行时执行指令
const help = `可用命令:
noop                           占位命令，输出 ok
adduser <username> <password>  新建用户
changepassword <username> <newpassword>  强制修改用户密码
`

func RunInteractiveShell(mode config.Mode) {
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
			case "changepassword":
				handleForceChangePassword(parts[1:])
			default:
				fmt.Print(help)
			}
		}
	}()
}

func handleNoop() {
	fmt.Println("ok")
}
