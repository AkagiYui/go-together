package repo

import (
	"time"
)

// User 用户表
type User struct {
	ID          int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string     `gorm:"column:name;type:varchar(255);not null" json:"name"`
	AccessKey   string     `gorm:"column:access_key;type:varchar(255);uniqueIndex;not null" json:"access_key"`
	IsActive    bool       `gorm:"column:is_active;not null;default:true" json:"is_active"`
	IsSuperuser bool       `gorm:"column:is_superuser;not null;default:false" json:"is_superuser"`
	CreatedAt   time.Time  `gorm:"column:created_at;type:timestamptz;not null;default:current_timestamp" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at;type:timestamptz" json:"updated_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// CreateUser 创建用户
func CreateUser(user User) (User, error) {
	result := DB.Create(&user)
	return user, result.Error
}

// GetUserByAccessKey 根据 access_key 获取用户
func GetUserByAccessKey(accessKey string) (User, error) {
	var user User
	result := DB.Where("access_key = ?", accessKey).First(&user)
	return user, result.Error
}

// GetUserByID 根据ID获取用户
func GetUserByID(id int64) (User, error) {
	var user User
	result := DB.First(&user, id)
	return user, result.Error
}

