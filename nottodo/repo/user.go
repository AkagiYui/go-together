package repo

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User 数据模型
// 注意：Password 仅用于输入，不会在查询中返回；对外响应不应包含密码。
type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	Nickname     string `json:"nickname"`
	Password     string `json:"password,omitempty"`
	PasswordHash string `json:"-"`
	CreatedAt    string `json:"created_at"`
}

func CreateUser(username, nickname, password string) (User, error) {
	if db == nil {
		return User{}, errors.New("db not initialized")
	}
	username = strings.TrimSpace(username)
	nickname = strings.TrimSpace(nickname)
	if username == "" || password == "" {
		return User{}, errors.New("用户名与密码不能为空")
	}
	if nickname == "" {
		nickname = username
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	var id int64
	var createdAt sql.NullTime
	err = db.QueryRow(`INSERT INTO users (username, nickname, password_hash) VALUES ($1, $2, $3) RETURNING id, created_at`,
		username, nickname, string(hash)).Scan(&id, &createdAt)
	if err != nil {
		return User{}, err
	}
	u := User{ID: id, Username: username, Nickname: nickname}
	if createdAt.Valid {
		u.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
	}
	return u, nil
}

func DeleteUserByID(id int64) (bool, error) {
	if db == nil {
		return false, errors.New("db not initialized")
	}
	res, err := db.Exec(`DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return false, err
	}
	affected, _ := res.RowsAffected()
	return affected > 0, nil
}

func GetUserByUsername(username string) (User, bool, error) {
	if db == nil {
		return User{}, false, errors.New("db not initialized")
	}
	row := db.QueryRow(`SELECT id, username, nickname, password_hash, created_at FROM users WHERE username = $1`, username)
	var u User
	var createdAt sql.NullTime
	if err := row.Scan(&u.ID, &u.Username, &u.Nickname, &u.PasswordHash, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, false, nil
		}
		return User{}, false, err
	}
	if createdAt.Valid {
		u.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
	}
	return u, true, nil
}

func UpdateUserPasswordByUsername(username, newPassword string) error {
	if db == nil {
		return errors.New("db not initialized")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec(`UPDATE users SET password_hash = $1, updated_at = $2 WHERE username = $3`, string(hash), time.Now(), username)
	return err
}
