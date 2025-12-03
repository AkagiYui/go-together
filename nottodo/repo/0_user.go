package repo

import (
	"time"

	"gorm.io/gorm"
)

// User 用户表
type User struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                                                                       // ID
	Username    string     `gorm:"column:username;type:varchar(255);uniqueIndex:idx_users_username;not null" json:"username"`                          // 用户名
	Password    string     `gorm:"column:password;type:varchar(255);not null" json:"password"`                                                         // 密码
	Nickname    *string    `gorm:"column:nickname;type:varchar(255)" json:"nickname"`                                                                  // 昵称（可空）
	RegisterAt  *time.Time `gorm:"column:register_at;type:timestamptz" json:"registerAt"`                                                             // 注册时间（可空）
	IsValidated bool       `gorm:"column:is_validated;not null;default:false" json:"isValidated"`                                                     // 是否已通过验证
	ValidatedAt *time.Time `gorm:"column:validated_at;type:timestamptz" json:"validatedAt"`                                                           // 验证时间（可空）
	CreatedAt   time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:current_timestamp;index:idx_users_created_at" json:"createdAt"` // 创建时间（非空）
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate GORM 钩子：创建用户前设置注册时间
func (u *User) BeforeCreate(_ *gorm.DB) error {
	if u.RegisterAt == nil {
		now := time.Now()
		u.RegisterAt = &now
	}
	return nil
}

// CreateUser 创建用户
func CreateUser(user User) (User, error) {
	result := DB.Create(&user)
	return user, result.Error
}

// GetUserByUsername 根据用户名获取用户
func GetUserByUsername(username string) (User, error) {
	var user User
	result := DB.Where("username = ?", username).First(&user)
	return user, result.Error
}

// GetUserByID 根据ID获取用户
func GetUserByID(id int64) (User, error) {
	var user User
	result := DB.First(&user, id)
	return user, result.Error
}

// UpdateUserPassword 更新用户密码
func UpdateUserPassword(userID int64, newPassword string) (User, error) {
	var user User
	result := DB.Model(&user).Where("id = ?", userID).Update("password", newPassword)
	if result.Error != nil {
		return user, result.Error
	}
	// 重新查询用户以返回完整信息
	result = DB.First(&user, userID)
	return user, result.Error
}
