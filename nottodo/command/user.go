package command

import (
	"fmt"

	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/nottodo/service/user"
)

func handleAddUser(args []string) {
	if len(args) != 2 {
		fmt.Println("用法: adduser <username> <password>")
		return
	}
	username := args[0]
	password := args[1]

	req := user.CreateUserRequest{Username: username, Password: password}
	if err := req.Validate(); err != nil {
		fmt.Println("错误: ", err)
		return
	}

	resp, err := req.Do()
	if err != nil {
		fmt.Println("错误: ", err)
		return
	}
	fmt.Printf("ok, 用户已创建，ID=%d, 用户名=%s\n", resp.(user.Response).ID, username)
}

func handleForceChangePassword(args []string) {
	if len(args) != 2 {
		fmt.Println("用法: changepassword <username> <newpassword>")
		return
	}
	username := args[0]
	newPassword := args[1]

	// 1. 获取用户 ID
	userRecord, err := repo.GetUserByUsername(username)
	if err != nil {
		fmt.Println("错误: ", err)
		return
	}

	// 2. 强制修改密码
	req := user.ForceChangePassword{UserID: userRecord.ID, NewPassword: newPassword}
	_, err = req.Do()
	if err != nil {
		fmt.Println("错误: ", err)
		return
	}
	fmt.Println("ok")
}
