package repo

func CreateUser(user User) (User, error) {
	return Db.CreateUser(Ctx, CreateUserParams{
		Username: user.Username,
		Password: user.Password,
		Nickname: user.Nickname,
	})
}
