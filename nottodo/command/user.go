package command

import (
	"fmt"

	usersvc "github.com/akagiyui/go-together/nottodo/service/user"
)

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

	resp, err := req.Do()
	if err != nil {
		fmt.Println("错误: ", err)
		return
	}
	fmt.Printf("ok, 用户已创建，ID=%d, 用户名=%s\n", resp.ID, username)
}
