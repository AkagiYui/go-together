package repo

func CreateEmail(email Email) (Email, error) {
	return Db.CreateEmail(Ctx, CreateEmailParams{
		UserID:     email.UserID,
		Email:      email.Email,
		IsPrimary:  email.IsPrimary,
		IsVerified: email.IsVerified,
	})
}

func GetEmailById(id int64) (Email, error) {
	return Db.GetEmail(Ctx, id)
}

func GetEmailByAddress(email string) (Email, error) {
	return Db.GetEmailByAddress(Ctx, email)
}

func ListEmailsByUserId(userID int64) ([]Email, error) {
	return Db.ListEmailsByUserId(Ctx, userID)
}

func ListEmails() ([]Email, error) {
	return Db.ListEmails(Ctx)
}

func UpdateEmail(email Email) error {
	return Db.UpdateEmail(Ctx, UpdateEmailParams{
		ID:         email.ID,
		IsPrimary:  email.IsPrimary,
		IsVerified: email.IsVerified,
	})
}
