package command

import (
	"fmt"

	"github.com/akagiyui/go-together/common/model"
	usersvc "github.com/akagiyui/go-together/nottodo/service/user"
	"github.com/akagiyui/go-together/rest"
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
