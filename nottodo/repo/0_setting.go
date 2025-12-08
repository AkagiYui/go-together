package repo

import (
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// Setting 系统设置表
type Setting struct {
	Key         string    `gorm:"column:key;primaryKey;type:varchar(255)" json:"key"`                         // 键
	Value       string    `gorm:"column:value;type:text;not null" json:"value"`                               // 值
	Description *string   `gorm:"column:description;type:text" json:"description"`                            // 描述（可空）
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamptz;not null;default:now()" json:"updatedAt"` // 更新时间（非空）
}

// TableName 指定表名
func (Setting) TableName() string {
	return "settings"
}

// GetIsAllowRegistration 获取是否允许注册的设置
func GetIsAllowRegistration() (bool, error) {
	var setting Setting
	result := DB.Where("key = ?", "is_allow_registration").First(&setting)

	// 如果记录不存在，返回默认值 false
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return false, nil
	}

	if result.Error != nil {
		return false, result.Error
	}

	return strconv.ParseBool(setting.Value)
}

// SetIsAllowRegistration 设置是否允许注册
func SetIsAllowRegistration(allowed bool) error {
	setting := Setting{
		Key:   "is_allow_registration",
		Value: strconv.FormatBool(allowed),
	}

	// 使用 GORM 的 Save 方法实现 UPSERT
	// Save 会根据主键判断是插入还是更新
	return DB.Save(&setting).Error
}
