// Package repo 提供数据库访问层功能
package repo

import "time"

// Email 邮箱表
type Email struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                           // ID
	UserID     int64     `gorm:"column:user_id;not null;index:idx_emails_user_id" json:"userId"`                         // 用户ID
	Email      string    `gorm:"column:email;type:varchar(255);uniqueIndex:idx_emails_email;not null" json:"email"`      // 邮箱地址
	IsPrimary  bool      `gorm:"column:is_primary;not null;default:false" json:"isPrimary"`                              // 是否为主要邮箱
	IsVerified bool      `gorm:"column:is_verified;not null;default:false" json:"isVerified"`                            // 是否已验证
	CreatedAt  time.Time `gorm:"column:created_at;type:timestamptz;not null;default:current_timestamp" json:"createdAt"` // 创建时间（非空）
}

// TableName 指定表名
func (Email) TableName() string {
	return "emails"
}

// CreateEmail 创建邮箱记录
func CreateEmail(email Email) (Email, error) {
	result := DB.Create(&email)
	return email, result.Error
}

// GetEmailByID 根据ID获取邮箱记录
func GetEmailByID(id int64) (Email, error) {
	var email Email
	result := DB.First(&email, id)
	return email, result.Error
}

// GetEmailByAddress 根据邮箱地址获取邮箱记录
func GetEmailByAddress(emailAddr string) (Email, error) {
	var email Email
	result := DB.Where("email = ?", emailAddr).First(&email)
	return email, result.Error
}

// ListEmailsByUserID 根据用户ID获取邮箱列表
func ListEmailsByUserID(userID int64) ([]Email, error) {
	var emails []Email
	result := DB.Where("user_id = ?", userID).Find(&emails)
	return emails, result.Error
}

// ListEmails 获取所有邮箱列表
func ListEmails() ([]Email, error) {
	var emails []Email
	result := DB.Find(&emails)
	return emails, result.Error
}

// UpdateEmail 更新邮箱记录
func UpdateEmail(email Email) error {
	return DB.Model(&Email{}).Where("id = ?", email.ID).Updates(map[string]any{
		"is_primary":  email.IsPrimary,
		"is_verified": email.IsVerified,
	}).Error
}
