package repo

func CreateUser(user User) (User, error) {
	return Db.CreateUser(Ctx, CreateUserParams{
		Username: user.Username,
		Password: user.Password,
		Nickname: user.Nickname,
	})
}

func GetUserByUsername(username string) (User, error) {
	return Db.GetUserByUsername(Ctx, username)
}

func GetUserById(id int64) (User, error) {
	return Db.GetUser(Ctx, id)
}
