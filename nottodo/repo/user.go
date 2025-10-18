package repo

import (
    "context"

    usergen "github.com/akagiyui/go-together/nottodo/repo/user"
)

// 兼容旧的仓储接口，内部转调 sqlc 生成的查询
// 注意：这里不再编写任何原生 SQL

type User = usergen.User

func CreateUser(ctx context.Context, username, nickname, password string) (User, error) {
    return UserQueries.CreateUser(ctx, username, nickname, password)
}

func DeleteUserByID(ctx context.Context, id int64) (bool, error) {
    affected, err := UserQueries.DeleteUser(ctx, id)
    if err != nil {
        return false, err
    }
    return affected > 0, nil
}

func GetUserByUsername(ctx context.Context, username string) (User, bool, error) {
    u, err := UserQueries.GetUserByUsername(ctx, username)
    if err != nil {
        return User{}, false, err
    }
    return u, true, nil
}

func UpdateUserPasswordByUsername(ctx context.Context, username, newPassword string) error {
    _, err := UserQueries.UpdateUserPassword(ctx, username, newPassword)
    return err
}
