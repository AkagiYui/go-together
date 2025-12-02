// Package encrypt 提供加密相关的功能
package encrypt

import (
	"crypto/sha256"
	"encoding/hex"
)

// Sha256 计算字节数组的 SHA256 哈希值
func Sha256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
