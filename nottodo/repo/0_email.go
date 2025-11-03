// Package repo 提供数据库访问层功能
package repo

// CreateEmail 创建邮箱记录
func CreateEmail(email Email) (Email, error) {
	return Db.CreateEmail(Ctx, CreateEmailParams{
		UserID:     email.UserID,
		Email:      email.Email,
		IsPrimary:  email.IsPrimary,
		IsVerified: email.IsVerified,
	})
}

// GetEmailByID 根据ID获取邮箱记录
func GetEmailByID(id int64) (Email, error) {
	return Db.GetEmail(Ctx, id)
}

// GetEmailByAddress 根据邮箱地址获取邮箱记录
func GetEmailByAddress(email string) (Email, error) {
	return Db.GetEmailByAddress(Ctx, email)
}

// ListEmailsByUserID 根据用户ID获取邮箱列表
func ListEmailsByUserID(userID int64) ([]Email, error) {
	return Db.ListEmailsByUserId(Ctx, userID)
}

// ListEmails 获取所有邮箱列表
func ListEmails() ([]Email, error) {
	return Db.ListEmails(Ctx)
}

// UpdateEmail 更新邮箱记录
func UpdateEmail(email Email) error {
	return Db.UpdateEmail(Ctx, UpdateEmailParams{
		ID:         email.ID,
		IsPrimary:  email.IsPrimary,
		IsVerified: email.IsVerified,
	})
}
