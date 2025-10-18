package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/akagiyui/go-together/nottodo/config"
)

// 在开发模式下启动交互式终端，支持在服务器运行时执行指令
// 当前仅提供一个占位命令：noop
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
			switch line {
			case "help", "h":
				fmt.Println("可用命令:")
				fmt.Println("  noop               占位命令，输出 ok")
			case "noop":
				fmt.Println("ok")
			default:
				fmt.Println("未知命令，输入 help 获取帮助")
			}
		}
	}()
}
