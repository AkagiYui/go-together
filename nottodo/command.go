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
const help = `可用命令:
noop               占位命令，输出 ok`

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
			switch line {
			case "noop":
				fmt.Println("ok")
			default:
				fmt.Println(help)
			}
		}
	}()
}
