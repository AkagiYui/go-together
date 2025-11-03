// Package user 提供用户认证相关的服务
package user

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/nottodo/cache"
	"github.com/akagiyui/go-together/nottodo/pkg"
	"github.com/akagiyui/go-together/nottodo/repo"
)

// GenerateTokenRequest 生成令牌请求
type GenerateTokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// GenerateTokenResponse 生成令牌响应
type GenerateTokenResponse struct {
	Token string `json:"token"`
}

// Do 执行生成令牌的业务逻辑
func (r GenerateTokenRequest) Do() (any, error) {
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

	// 将 token 和用户 ID 写入缓存
	err = cache.Set("auth_token:"+token, user.ID, 24*time.Hour)
	if err != nil {
		return "", err
	}

	return GenerateTokenResponse{Token: token}, nil
}
