package user

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/common/validation"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/argon2"
)

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

func (r *CreateUserRequest) Validate() error {
	return errors.Join(
		validation.Required(r.Username, "用户名"),
		validation.Required(r.Password, "密码"),
	)
}

type UserResponse struct {
	ID int64 `json:"id"`
}

func NewUserResponse(user repo.User) UserResponse {
	return UserResponse{
		ID: user.ID,
	}
}

func (r *CreateUserRequest) Handle(ctx *rest.Context) {
	password, err := hashPassword(r.Password)
	if err != nil {
		ctx.SetResult(model.InternalError(err))
		return
	}

	r.Password = password
	newUser, err := repo.CreateUser(repo.User{
		Username: r.Username,
		Password: r.Password,
		Nickname: pgtype.Text{String: r.Nickname, Valid: r.Nickname != ""},
	})
	if err != nil {
		ctx.SetResult(model.InternalError(err))
		return
	}
	ctx.SetResult(model.Success(NewUserResponse(newUser)))
}

func hashPassword(password string) (string, error) {
	// 推荐的安全参数
	params := &struct {
		memory      uint32
		iterations  uint32
		parallelism uint8
		saltLength  uint32
		keyLength   uint32
	}{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	}

	// 生成随机盐
	salt := make([]byte, params.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// 使用 argon2.IDKey 生成哈希
	hash := argon2.IDKey([]byte(password), salt, params.iterations, params.memory, params.parallelism, params.keyLength)

	// 将盐和哈希编码为 Base64 字符串
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// 返回符合 PHC 字符串格式的哈希字符串，便于存储和验证
	// $argon2id$v=19$m=65536,t=3,p=2$salt$hash
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		params.memory,
		params.iterations,
		params.parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

func verifyPassword(password, encodedHash string) (bool, error) {
	// 分割哈希字符串
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid hash format")
	}

	// 解析参数
	var version int
	var memory, iterations uint32
	var parallelism uint8
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}
	if version != argon2.Version {
		return false, fmt.Errorf("incompatible version")
	}

	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, err
	}

	// 解码盐和哈希
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}
	keyLength := uint32(len(hash))

	// 使用相同的参数重新生成哈希
	otherHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)

	// 使用固定时间比较来防止时序攻击
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}
