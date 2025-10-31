package user

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/nottodo/pkg"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

type GenerateTokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type GenerateTokenResponse struct {
	Token string `json:"token"`
}

func (r GenerateTokenRequest) Handle(ctx *rest.Context) {
	token, err := r.Do()
	if err != nil {
		ctx.SetResult(model.Error(err))
		return
	}
	ctx.SetResult(model.Success(GenerateTokenResponse{Token: token}))
}

func (r GenerateTokenRequest) Do() (string, error) {
	// 校验用户名和密码
	user, err := repo.GetUserByUsername(r.Username)
	if err != nil {
		return "", err
	}

	match, err := pkg.VerifyPassword(r.Password, user.Password)
	if err != nil {
		return "", err
	}
	if !match {
		return "", model.ErrUnauthorized
	}

	// 生成 32 字节的随机数据
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	// 将随机字节转换为十六进制字符串
	// 结果是 64 个字符的字符串（32 字节 * 2）
	token := hex.EncodeToString(tokenBytes)

	return token, nil
}
