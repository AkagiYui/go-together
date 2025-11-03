package repo

// CreateUser 创建用户
func CreateUser(user User) (User, error) {
	return Db.CreateUser(Ctx, CreateUserParams{
		Username: user.Username,
		Password: user.Password,
		Nickname: user.Nickname,
	})
}

// GetUserByUsername 根据用户名获取用户
func GetUserByUsername(username string) (User, error) {
	return Db.GetUserByUsername(Ctx, username)
}

// GetUserByID 根据ID获取用户
func GetUserByID(id int64) (User, error) {
	return Db.GetUser(Ctx, id)
}

// UpdateUserPassword 更新用户密码
func UpdateUserPassword(userID int64, newPassword string) (User, error) {
	return Db.UpdateUserPassword(Ctx, UpdateUserPasswordParams{
		ID:       userID,
		Password: newPassword,
	})
}
