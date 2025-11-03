package command

import (
	"fmt"

	"github.com/akagiyui/go-together/nottodo/repo"
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

func handleForceChangePassword(args []string) {
	if len(args) != 2 {
		fmt.Println("用法: changepassword <username> <newpassword>")
		return
	}
	username := args[0]
	newPassword := args[1]

	// 1. 获取用户 ID
	user, err := repo.GetUserByUsername(username)
	if err != nil {
		fmt.Println("错误: ", err)
		return
	}

	// 2. 强制修改密码
	req := usersvc.ForceChangePassword{UserID: user.ID, NewPassword: newPassword}
	_, err = req.Do()
	if err != nil {
		fmt.Println("错误: ", err)
		return
	}
	fmt.Println("ok")
}
