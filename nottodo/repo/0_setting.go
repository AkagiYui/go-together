package repo

import (
	"database/sql"
	"errors"
	"strconv"
)

// GetIsAllowRegistration 获取是否允许注册的设置
func GetIsAllowRegistration() (bool, error) {
	setting, err := Db.GetSetting(Ctx, "is_allow_registration")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return strconv.ParseBool(setting.Value)
}

// SetIsAllowRegistration 设置是否允许注册
func SetIsAllowRegistration(allowed bool) error {
	_, err := Db.SetSetting(Ctx, SetSettingParams{
		Key:   "is_allow_registration",
		Value: strconv.FormatBool(allowed),
	})
	return err
}
