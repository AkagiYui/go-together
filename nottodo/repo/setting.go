package repo

import (
	"database/sql"
	"errors"
	"strconv"
)

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

func SetIsAllowRegistration(allowed bool) error {
	_, err := Db.SetSetting(Ctx, SetSettingParams{
		Key:   "is_allow_registration",
		Value: strconv.FormatBool(allowed),
	})
	return err
}
